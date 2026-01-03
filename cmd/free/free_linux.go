//go:build linux

package free

import (
	"fmt"
	"io"

	"github.com/shirou/gopsutil/v3/mem"
)

func printMemoryInfo(w io.Writer, virtualMem *mem.VirtualMemoryStat, swapMem *mem.SwapMemoryStat, unitFactor float64, unitLabel string) {
	// Linux format with shared, buff/cache columns
	fmt.Fprintf(w, "%12s %10s %10s %10s %10s %10s %10s\n", "", "total", "used", "free", "shared", "buff/cache", "available")
	fmt.Fprintf(w, "%12s %10.0f %10.0f %10.0f %10.0f %10.0f %10.0f %s\n",
		"Mem:",
		float64(virtualMem.Total)/unitFactor,
		float64(virtualMem.Used)/unitFactor,
		float64(virtualMem.Free)/unitFactor,
		float64(virtualMem.Shared)/unitFactor,
		float64(virtualMem.Buffers+virtualMem.Cached)/unitFactor,
		float64(virtualMem.Available)/unitFactor,
		unitLabel,
	)
	fmt.Fprintf(w, "%12s %10.0f %10.0f %10.0f %s\n",
		"Swap:",
		float64(swapMem.Total)/unitFactor,
		float64(swapMem.Used)/unitFactor,
		float64(swapMem.Free)/unitFactor,
		unitLabel,
	)
}
