//go:build (linux && cgo) || windows || darwin

package jukebox

import (
	"bytes"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

// AudioAvailable indicates whether audio playback is supported in this build.
const AudioAvailable = true

// player handles the actual audio output using beep.
type player struct {
	mu sync.Mutex

	initialized  bool
	sampleRate   beep.SampleRate
	ctrl         *beep.Ctrl
	streamer     beep.StreamSeekCloser
	format       beep.Format
	resampled    beep.Streamer
	done         chan struct{}
	onSongDone   func() // Callback when song finishes
}

// newPlayer creates a new audio player.
func newPlayer() *player {
	return &player{
		sampleRate: beep.SampleRate(44100), // Standard sample rate
	}
}

// initSpeaker initializes the speaker if not already done.
func (p *player) initSpeaker(format beep.Format) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return nil
	}

	err := speaker.Init(p.sampleRate, p.sampleRate.N(time.Second/10))
	if err != nil {
		return err
	}
	p.initialized = true
	return nil
}

// playSong starts playing a song from its in-memory data.
func (p *player) playSong(song *Song, onDone func()) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Stop any current playback
	p.stopLocked()

	// Decode MP3 from memory
	reader := bytes.NewReader(song.Data)
	streamer, format, err := mp3.Decode(nopCloser{reader})
	if err != nil {
		return err
	}

	p.streamer = streamer
	p.format = format
	p.onSongDone = onDone

	// Initialize speaker with first song's sample rate if needed
	if !p.initialized {
		p.mu.Unlock()
		err = p.initSpeaker(format)
		p.mu.Lock()
		if err != nil {
			streamer.Close()
			return err
		}
	}

	// Resample if needed to match speaker sample rate
	p.resampled = beep.Resample(4, format.SampleRate, p.sampleRate, streamer)

	// Create control for pause/resume
	p.ctrl = &beep.Ctrl{Streamer: p.resampled, Paused: false}

	// Create done channel
	p.done = make(chan struct{})

	// Play with callback
	speaker.Play(beep.Seq(p.ctrl, beep.Callback(func() {
		close(p.done)
		if p.onSongDone != nil {
			// Run callback in separate goroutine to avoid deadlock
			// when the callback tries to play the next song
			go p.onSongDone()
		}
	})))

	return nil
}

// pause pauses playback.
func (p *player) pause() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ctrl != nil {
		speaker.Lock()
		p.ctrl.Paused = true
		speaker.Unlock()
	}
}

// resume resumes playback.
func (p *player) resume() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ctrl != nil {
		speaker.Lock()
		p.ctrl.Paused = false
		speaker.Unlock()
	}
}

// stop stops playback completely.
func (p *player) stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopLocked()
}

// stopLocked stops playback (must be called with lock held).
func (p *player) stopLocked() {
	if p.ctrl != nil {
		speaker.Lock()
		p.ctrl.Paused = true
		speaker.Unlock()
	}
	if p.streamer != nil {
		p.streamer.Close()
		p.streamer = nil
	}
	p.ctrl = nil
	p.resampled = nil
	p.done = nil
}

// position returns the current playback position.
func (p *player) position() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.streamer == nil {
		return 0
	}

	speaker.Lock()
	pos := p.streamer.Position()
	speaker.Unlock()

	return p.format.SampleRate.D(pos)
}

// seek sets the playback position.
func (p *player) seek(d time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.streamer == nil {
		return nil
	}

	speaker.Lock()
	defer speaker.Unlock()

	samples := p.format.SampleRate.N(d)
	return p.streamer.Seek(samples)
}

// duration returns the total duration of the current song.
func (p *player) duration() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.streamer == nil {
		return 0
	}

	return p.format.SampleRate.D(p.streamer.Len())
}

// isPaused returns whether playback is paused.
func (p *player) isPaused() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ctrl == nil {
		return false
	}

	speaker.Lock()
	paused := p.ctrl.Paused
	speaker.Unlock()
	return paused
}

// nopCloser wraps a bytes.Reader to implement io.ReadCloser.
type nopCloser struct {
	*bytes.Reader
}

func (nopCloser) Close() error { return nil }
