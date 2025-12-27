//go:build !cgo

package bird

import "embed"

// musicFS is empty in non-CGO builds since audio playback is unavailable.
var musicFS embed.FS
