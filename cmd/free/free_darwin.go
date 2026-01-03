//go:build darwin

package free

import (
	"fmt"
	"io"

	"github.com/shirou/gopsutil/v3/mem"
)

func printMemoryInfo(w io.Writer, virtualMem *mem.VirtualMemoryStat, swapMem *mem.SwapMemoryStat, unitFactor float64, unitLabel string) {
	// macOS format: show Active, Inactive, Wired, Free (macOS memory categories)
	// These are more meaningful on macOS than Linux's shared/buff/cache

	fmt.Fprintf(w, "%12s %10s %10s %10s %10s %10s %10s\n", "", "total", "used", "active", "inactive", "wired", "available")
	fmt.Fprintf(w, "%12s %10.0f %10.0f %10.0f %10.0f %10.0f %10.0f %s\n",
		"Mem:",
		float64(virtualMem.Total)/unitFactor,
		float64(virtualMem.Used)/unitFactor,
		float64(virtualMem.Active)/unitFactor,
		float64(virtualMem.Inactive)/unitFactor,
		float64(virtualMem.Wired)/unitFactor,
		float64(virtualMem.Available)/unitFactor,
		unitLabel,
	)

	// Swap
	if swapMem.Total > 0 {
		fmt.Fprintf(w, "%12s %10.0f %10.0f %10.0f %s\n",
			"Swap:",
			float64(swapMem.Total)/unitFactor,
			float64(swapMem.Used)/unitFactor,
			float64(swapMem.Free)/unitFactor,
			unitLabel,
		)
	} else {
		fmt.Fprintf(w, "%12s %10s\n", "Swap:", "(disabled)")
	}
}
