package jukebox

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	ErrSongNotFound    = errors.New("song not found")
	ErrNoSongsLoaded   = errors.New("no songs loaded")
	ErrQueueEmpty      = errors.New("queue is empty")
	ErrInvalidFile     = errors.New("invalid file: must be an MP3")
	ErrAlreadyPlaying  = errors.New("already playing")
	ErrNotPlaying      = errors.New("not currently playing")
	ErrIndexOutOfRange = errors.New("index out of range")
)

// Jukebox manages a collection of songs in memory with playback controls.
type Jukebox struct {
	mu sync.RWMutex

	// Song storage
	songs map[string]*Song // All loaded songs by ID

	// Queue management
	queue        []*Song // Current play queue
	shuffleQueue []*Song // Shuffled version of the queue
	queueIndex   int     // Current position in queue (-1 if not set)

	// Playback state
	state      PlaybackState
	shuffle    bool
	playbackID uint64 // Incremented each time a new song starts, used to ignore stale callbacks

	// Audio player
	player *player
}

// New creates a new Jukebox instance.
func New() *Jukebox {
	return &Jukebox{
		songs:      make(map[string]*Song),
		queue:      make([]*Song, 0),
		queueIndex: -1,
		state:      StateStopped,
		shuffle:    false,
		player:     newPlayer(),
	}
}

// generateID creates a unique identifier for a song.
func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Load reads an MP3 file from disk into memory and adds it to the jukebox.
// Returns the Song with its assigned ID.
func (j *Jukebox) Load(path string) (*Song, error) {
	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(path), ".mp3") {
		return nil, ErrInvalidFile
	}

	// Read file into memory
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Create song
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	song := &Song{
		ID:       generateID(),
		Name:     name,
		Path:     path,
		Data:     data,
		Size:     info.Size(),
		LoadedAt: time.Now(),
	}

	// Add to jukebox
	j.mu.Lock()
	j.songs[song.ID] = song
	j.queue = append(j.queue, song)
	j.mu.Unlock()

	return song, nil
}

// LoadBytes loads raw MP3 data directly into the jukebox.
// Useful for loading from sources other than local files.
func (j *Jukebox) LoadBytes(name string, data []byte) (*Song, error) {
	if len(data) == 0 {
		return nil, ErrInvalidFile
	}

	song := &Song{
		ID:       generateID(),
		Name:     name,
		Path:     "",
		Data:     data,
		Size:     int64(len(data)),
		LoadedAt: time.Now(),
	}

	j.mu.Lock()
	j.songs[song.ID] = song
	j.queue = append(j.queue, song)
	j.mu.Unlock()

	return song, nil
}

// Remove removes a song from the jukebox by ID, freeing its memory.
func (j *Jukebox) Remove(id string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	song, exists := j.songs[id]
	if !exists {
		return ErrSongNotFound
	}

	// Remove from songs map
	delete(j.songs, id)

	// Remove from queue
	for i, s := range j.queue {
		if s.ID == id {
			j.queue = append(j.queue[:i], j.queue[i+1:]...)
			// Adjust queue index if needed
			if j.queueIndex >= i && j.queueIndex > 0 {
				j.queueIndex--
			}
			break
		}
	}

	// Remove from shuffle queue if it exists
	for i, s := range j.shuffleQueue {
		if s.ID == id {
			j.shuffleQueue = append(j.shuffleQueue[:i], j.shuffleQueue[i+1:]...)
			break
		}
	}

	// Clear the data to help GC
	song.Data = nil

	// If we removed the current song, stop playback
	if j.queueIndex >= len(j.queue) {
		j.queueIndex = -1
		j.state = StateStopped
		j.player.stop()
	}

	return nil
}

// Clear removes all songs from the jukebox and stops playback.
func (j *Jukebox) Clear() {
	j.mu.Lock()
	defer j.mu.Unlock()

	// Stop audio playback
	j.player.stop()

	// Clear data to help GC
	for _, song := range j.songs {
		song.Data = nil
	}

	j.songs = make(map[string]*Song)
	j.queue = make([]*Song, 0)
	j.shuffleQueue = nil
	j.queueIndex = -1
	j.state = StateStopped
}

// Get returns a song by ID.
func (j *Jukebox) Get(id string) (*Song, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	song, exists := j.songs[id]
	if !exists {
		return nil, ErrSongNotFound
	}
	return song, nil
}

// List returns all loaded songs.
func (j *Jukebox) List() []*Song {
	j.mu.RLock()
	defer j.mu.RUnlock()

	songs := make([]*Song, 0, len(j.songs))
	for _, song := range j.songs {
		songs = append(songs, song)
	}
	return songs
}

// Queue returns the current play queue.
func (j *Jukebox) Queue() []*Song {
	j.mu.RLock()
	defer j.mu.RUnlock()

	return j.currentQueue()
}

// currentQueue returns the active queue (shuffled or normal).
// Must be called with at least a read lock held.
func (j *Jukebox) currentQueue() []*Song {
	if j.shuffle && j.shuffleQueue != nil {
		return j.shuffleQueue
	}
	return j.queue
}

// Count returns the number of loaded songs.
func (j *Jukebox) Count() int {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return len(j.songs)
}

// TotalSize returns the total size of all loaded songs in bytes.
func (j *Jukebox) TotalSize() int64 {
	j.mu.RLock()
	defer j.mu.RUnlock()

	var total int64
	for _, song := range j.songs {
		total += song.Size
	}
	return total
}
