package jukebox

import (
	"time"
)

// Song represents an MP3 file loaded into memory.
type Song struct {
	ID       string        // Unique identifier
	Name     string        // Display name (typically filename without extension)
	Path     string        // Original file path (for reference only)
	Data     []byte        // Raw MP3 data in memory
	Duration time.Duration // Duration of the song (if extractable)
	Size     int64         // Size in bytes
	LoadedAt time.Time     // When the song was loaded
}

// PlaybackState represents the current state of playback.
type PlaybackState string

const (
	StateStopped PlaybackState = "stopped"
	StatePlaying PlaybackState = "playing"
	StatePaused  PlaybackState = "paused"
)

// PlaybackInfo contains information about the current playback state.
type PlaybackInfo struct {
	State       PlaybackState
	CurrentSong *Song
	Position    time.Duration // Current position in the song
	QueueLength int           // Number of songs in the queue
	QueueIndex  int           // Current position in the queue (-1 if not playing)
	Shuffle     bool          // Whether shuffle is enabled
}
