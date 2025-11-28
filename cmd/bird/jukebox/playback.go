package jukebox

import (
	"math/rand"
	"time"
)

// Play starts or resumes playback.
// If no song is selected, starts from the beginning of the queue.
func (j *Jukebox) Play() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	queue := j.currentQueue()
	if len(queue) == 0 {
		return ErrQueueEmpty
	}

	// If paused, just resume
	if j.state == StatePaused {
		j.player.resume()
		j.state = StatePlaying
		return nil
	}

	if j.state == StatePlaying {
		return ErrAlreadyPlaying
	}

	// If stopped, start from beginning or current position
	if j.queueIndex < 0 {
		j.queueIndex = 0
	}

	// Start playing the current song
	song := queue[j.queueIndex]
	j.playbackID++
	currentID := j.playbackID
	err := j.player.playSong(song, func() { j.onSongFinished(currentID) })
	if err != nil {
		return err
	}

	j.state = StatePlaying
	return nil
}

// onSongFinished is called when a song finishes playing.
// The id parameter is used to ignore stale callbacks from songs that were
// skipped or replaced by manual user actions (e.g., Next, Previous, PlaySong).
func (j *Jukebox) onSongFinished(id uint64) {
	j.mu.Lock()
	defer j.mu.Unlock()

	// Ignore stale callbacks from previous playback sessions
	if id != j.playbackID {
		return
	}

	queue := j.currentQueue()
	if len(queue) == 0 {
		j.state = StateStopped
		return
	}

	// Advance to next song
	j.queueIndex++
	if j.queueIndex >= len(queue) {
		j.queueIndex = 0 // Loop back to beginning
	}

	// Play next song
	song := queue[j.queueIndex]
	j.playbackID++
	currentID := j.playbackID
	err := j.player.playSong(song, func() { j.onSongFinished(currentID) })
	if err != nil {
		j.state = StateStopped
		return
	}
}

// PlaySong starts playing a specific song by ID.
func (j *Jukebox) PlaySong(id string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	// Find the song in the queue
	queue := j.currentQueue()
	for i, song := range queue {
		if song.ID == id {
			j.queueIndex = i
			j.playbackID++
			currentID := j.playbackID
			err := j.player.playSong(song, func() { j.onSongFinished(currentID) })
			if err != nil {
				return err
			}
			j.state = StatePlaying
			return nil
		}
	}

	return ErrSongNotFound
}

// PlayIndex starts playing a song at a specific queue index.
func (j *Jukebox) PlayIndex(index int) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	queue := j.currentQueue()
	if index < 0 || index >= len(queue) {
		return ErrIndexOutOfRange
	}

	j.queueIndex = index
	song := queue[j.queueIndex]
	j.playbackID++
	currentID := j.playbackID
	err := j.player.playSong(song, func() { j.onSongFinished(currentID) })
	if err != nil {
		return err
	}
	j.state = StatePlaying
	return nil
}

// Pause pauses playback.
func (j *Jukebox) Pause() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.state != StatePlaying {
		return ErrNotPlaying
	}

	j.player.pause()
	j.state = StatePaused
	return nil
}

// Stop stops playback and resets position.
func (j *Jukebox) Stop() {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.player.stop()
	j.state = StateStopped
}

// Next advances to the next song in the queue.
func (j *Jukebox) Next() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	queue := j.currentQueue()
	if len(queue) == 0 {
		return ErrQueueEmpty
	}

	j.queueIndex++
	if j.queueIndex >= len(queue) {
		j.queueIndex = 0 // Loop back to beginning
	}

	// If we were playing, start the new song
	if j.state == StatePlaying {
		song := queue[j.queueIndex]
		j.playbackID++
		currentID := j.playbackID
		err := j.player.playSong(song, func() { j.onSongFinished(currentID) })
		if err != nil {
			j.state = StateStopped
			return err
		}
	}

	return nil
}

// Previous goes back to the previous song in the queue.
func (j *Jukebox) Previous() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	queue := j.currentQueue()
	if len(queue) == 0 {
		return ErrQueueEmpty
	}

	j.queueIndex--
	if j.queueIndex < 0 {
		j.queueIndex = len(queue) - 1 // Loop to end
	}

	// If we were playing, start the new song
	if j.state == StatePlaying {
		song := queue[j.queueIndex]
		j.playbackID++
		currentID := j.playbackID
		err := j.player.playSong(song, func() { j.onSongFinished(currentID) })
		if err != nil {
			j.state = StateStopped
			return err
		}
	}

	return nil
}

// Seek sets the playback position.
func (j *Jukebox) Seek(position time.Duration) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	if position < 0 {
		position = 0
	}
	return j.player.seek(position)
}

// Position returns the current playback position.
func (j *Jukebox) Position() time.Duration {
	return j.player.position()
}

// Duration returns the duration of the current song.
func (j *Jukebox) Duration() time.Duration {
	return j.player.duration()
}

// SetShuffle enables or disables shuffle mode.
func (j *Jukebox) SetShuffle(enabled bool) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if enabled == j.shuffle {
		return
	}

	j.shuffle = enabled

	if enabled {
		j.shuffleQueue = j.createShuffledQueue()
		// Find current song in new shuffle queue
		if j.queueIndex >= 0 && j.queueIndex < len(j.queue) {
			currentSong := j.queue[j.queueIndex]
			for i, song := range j.shuffleQueue {
				if song.ID == currentSong.ID {
					j.queueIndex = i
					break
				}
			}
		}
	} else {
		// Switch back to normal queue
		if j.queueIndex >= 0 && j.shuffleQueue != nil && j.queueIndex < len(j.shuffleQueue) {
			currentSong := j.shuffleQueue[j.queueIndex]
			for i, song := range j.queue {
				if song.ID == currentSong.ID {
					j.queueIndex = i
					break
				}
			}
		}
		j.shuffleQueue = nil
	}
}

// ToggleShuffle toggles shuffle mode and returns the new state.
func (j *Jukebox) ToggleShuffle() bool {
	j.mu.Lock()
	newState := !j.shuffle
	j.mu.Unlock()

	j.SetShuffle(newState)
	return newState
}

// Reshuffle creates a new shuffle order.
func (j *Jukebox) Reshuffle() {
	j.mu.Lock()
	defer j.mu.Unlock()

	if !j.shuffle {
		return
	}

	var currentSong *Song
	if j.queueIndex >= 0 && j.shuffleQueue != nil && j.queueIndex < len(j.shuffleQueue) {
		currentSong = j.shuffleQueue[j.queueIndex]
	}

	j.shuffleQueue = j.createShuffledQueue()

	// Maintain current song position
	if currentSong != nil {
		for i, song := range j.shuffleQueue {
			if song.ID == currentSong.ID {
				j.queueIndex = i
				break
			}
		}
	}
}

// createShuffledQueue creates a shuffled copy of the queue.
// Must be called with lock held.
func (j *Jukebox) createShuffledQueue() []*Song {
	shuffled := make([]*Song, len(j.queue))
	copy(shuffled, j.queue)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(shuffled), func(i, k int) {
		shuffled[i], shuffled[k] = shuffled[k], shuffled[i]
	})

	return shuffled
}

// Status returns the current playback information.
func (j *Jukebox) Status() PlaybackInfo {
	j.mu.RLock()
	defer j.mu.RUnlock()

	queue := j.currentQueue()
	info := PlaybackInfo{
		State:       j.state,
		Position:    j.player.position(),
		QueueLength: len(queue),
		QueueIndex:  j.queueIndex,
		Shuffle:     j.shuffle,
	}

	if j.queueIndex >= 0 && j.queueIndex < len(queue) {
		info.CurrentSong = queue[j.queueIndex]
	}

	return info
}

// CurrentSong returns the currently playing/paused song, or nil if none.
func (j *Jukebox) CurrentSong() *Song {
	j.mu.RLock()
	defer j.mu.RUnlock()

	queue := j.currentQueue()
	if j.queueIndex >= 0 && j.queueIndex < len(queue) {
		return queue[j.queueIndex]
	}
	return nil
}

// IsPlaying returns true if a song is currently playing.
func (j *Jukebox) IsPlaying() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.state == StatePlaying
}

// IsPaused returns true if playback is paused.
func (j *Jukebox) IsPaused() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.state == StatePaused
}

// IsStopped returns true if playback is stopped.
func (j *Jukebox) IsStopped() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.state == StateStopped
}

// IsShuffled returns true if shuffle mode is enabled.
func (j *Jukebox) IsShuffled() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.shuffle
}
