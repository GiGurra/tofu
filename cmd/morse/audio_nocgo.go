//go:build linux && !cgo

package morse

import (
	"fmt"
	"time"
)

func playMorse(morse string, wpm int) {
	// Calculate timing based on WPM
	unitDuration := time.Duration(float64(time.Second) * 60 / (50 * float64(wpm)))
	dotDuration := unitDuration
	dashDuration := 3 * unitDuration
	elementGap := unitDuration
	letterGap := 3 * unitDuration
	wordGap := 7 * unitDuration

	// Fallback: use terminal bell and visual output
	fmt.Println("(Audio requires CGO on Linux. Using terminal bell...)")

	for i, char := range morse {
		switch char {
		case '.':
			fmt.Print("\a") // Terminal bell
			time.Sleep(dotDuration + elementGap)
		case '-':
			fmt.Print("\a")
			time.Sleep(dashDuration + elementGap)
		case ' ':
			if i+1 < len(morse) && morse[i+1] == '/' {
				continue
			}
			if i > 0 && morse[i-1] == '/' {
				continue
			}
			time.Sleep(letterGap - elementGap)
		case '/':
			time.Sleep(wordGap - letterGap)
		}
	}
}
