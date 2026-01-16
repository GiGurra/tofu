//go:build (linux && cgo) || windows || darwin

package morse

import (
	"math"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

const (
	sampleRate = 44100
	frequency  = 700 // Hz - standard morse tone
)

var speakerInitialized = false

func playMorse(morse string, wpm int) {
	// Calculate timing based on WPM
	// PARIS = 50 units, so unit duration = 60 / (50 * WPM) seconds
	unitDuration := time.Duration(float64(time.Second) * 60 / (50 * float64(wpm)))
	dotDuration := unitDuration
	dashDuration := 3 * unitDuration
	elementGap := unitDuration
	letterGap := 3 * unitDuration
	wordGap := 7 * unitDuration

	// Initialize speaker
	if !speakerInitialized {
		err := speaker.Init(beep.SampleRate(sampleRate), sampleRate/10)
		if err != nil {
			return
		}
		speakerInitialized = true
	}

	for i, char := range morse {
		switch char {
		case '.':
			playTone(dotDuration)
			time.Sleep(elementGap)
		case '-':
			playTone(dashDuration)
			time.Sleep(elementGap)
		case ' ':
			// Check if it's a word separator (/) or letter separator
			if i+1 < len(morse) && morse[i+1] == '/' {
				continue // Skip, will handle with /
			}
			if i > 0 && morse[i-1] == '/' {
				continue // Skip, already handled
			}
			time.Sleep(letterGap - elementGap) // Already waited elementGap
		case '/':
			time.Sleep(wordGap - letterGap) // Already waited letterGap worth
		}
	}
}

func playTone(duration time.Duration) {
	samples := int(float64(sampleRate) * duration.Seconds())

	streamer := &toneStreamer{
		samples:   samples,
		position:  0,
		frequency: frequency,
	}

	done := make(chan struct{})
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		close(done)
	})))
	<-done
}

type toneStreamer struct {
	samples   int
	position  int
	frequency float64
}

func (t *toneStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	for i := range samples {
		if t.position >= t.samples {
			return i, false
		}

		// Generate sine wave with envelope to avoid clicks
		phase := 2 * math.Pi * t.frequency * float64(t.position) / float64(sampleRate)
		value := math.Sin(phase)

		// Apply envelope (fade in/out)
		envelope := 1.0
		fadeLen := t.samples / 20 // 5% fade
		if fadeLen < 10 {
			fadeLen = 10
		}
		if t.position < fadeLen {
			envelope = float64(t.position) / float64(fadeLen)
		} else if t.position > t.samples-fadeLen {
			envelope = float64(t.samples-t.position) / float64(fadeLen)
		}

		value *= envelope * 0.5 // 50% volume
		samples[i][0] = value
		samples[i][1] = value
		t.position++
	}
	return len(samples), true
}

func (t *toneStreamer) Err() error {
	return nil
}
