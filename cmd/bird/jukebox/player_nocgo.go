//go:build !cgo

package jukebox

import (
	"time"
)

// AudioAvailable indicates whether audio playback is supported in this build.
// Audio requires CGO for native sound libraries.
const AudioAvailable = false

// player is a no-op audio player for builds without cgo.
// The game will work but without sound.
type player struct{}

// newPlayer creates a new no-op player.
func newPlayer() *player {
	return &player{}
}

// playSong is a no-op when cgo is disabled.
func (p *player) playSong(song *Song, onDone func()) error {
	return nil
}

// pause is a no-op when cgo is disabled.
func (p *player) pause() {}

// resume is a no-op when cgo is disabled.
func (p *player) resume() {}

// stop is a no-op when cgo is disabled.
func (p *player) stop() {}

// position returns 0 when cgo is disabled.
func (p *player) position() time.Duration {
	return 0
}

// seek is a no-op when cgo is disabled.
func (p *player) seek(d time.Duration) error {
	return nil
}

// duration returns 0 when cgo is disabled.
func (p *player) duration() time.Duration {
	return 0
}

// isPaused returns false when cgo is disabled.
func (p *player) isPaused() bool {
	return false
}
