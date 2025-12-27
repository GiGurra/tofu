//go:build (linux && cgo) || windows || darwin

package bird

import "embed"

// musicFS embeds the music directory for audio playback.
//
//go:embed music
var musicFS embed.FS
