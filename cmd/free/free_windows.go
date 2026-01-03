//go:build windows

package free

import (
	"fmt"
	"io"
	"syscall"
	"unsafe"

	"github.com/shirou/gopsutil/v3/mem"
)

// MEMORYSTATUSEX for getting commit charge info
type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

func getCommitInfo() (commitTotal, commitLimit uint64, err error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	globalMemoryStatusEx := kernel32.NewProc("GlobalMemoryStatusEx")

	var memStatus memoryStatusEx
	memStatus.Length = uint32(unsafe.Sizeof(memStatus))

	ret, _, err := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memStatus)))
	if ret == 0 {
		return 0, 0, err
	}

	// Commit limit = TotalPageFile (which includes physical memory + page file)
	// Commit total (in use) = TotalPageFile - AvailPageFile
	commitLimit = memStatus.TotalPageFile
	commitTotal = memStatus.TotalPageFile - memStatus.AvailPageFile
	return commitTotal, commitLimit, nil
}

func printMemoryInfo(w io.Writer, virtualMem *mem.VirtualMemoryStat, swapMem *mem.SwapMemoryStat, unitFactor float64, unitLabel string) {
	// Windows format: show use% instead of shared/buff/cache (which are 0 on Windows)
	// Also show commit charge which is useful on Windows

	// Get commit info
	commitUsed, commitLimit, _ := getCommitInfo()

	fmt.Fprintf(w, "%12s %10s %10s %10s %10s\n", "", "total", "used", "available", "use%")
	fmt.Fprintf(w, "%12s %10.0f %10.0f %10.0f %9.1f%% %s\n",
		"Physical:",
		float64(virtualMem.Total)/unitFactor,
		float64(virtualMem.Used)/unitFactor,
		float64(virtualMem.Available)/unitFactor,
		virtualMem.UsedPercent,
		unitLabel,
	)

	// Commit charge (virtual memory committed by processes)
	if commitLimit > 0 {
		commitPercent := float64(commitUsed) / float64(commitLimit) * 100
		fmt.Fprintf(w, "%12s %10.0f %10.0f %10.0f %9.1f%% %s\n",
			"Commit:",
			float64(commitLimit)/unitFactor,
			float64(commitUsed)/unitFactor,
			float64(commitLimit-commitUsed)/unitFactor,
			commitPercent,
			unitLabel,
		)
	}

	// Page file (swap)
	if swapMem.Total > 0 {
		swapPercent := float64(0)
		if swapMem.Total > 0 {
			swapPercent = float64(swapMem.Used) / float64(swapMem.Total) * 100
		}
		fmt.Fprintf(w, "%12s %10.0f %10.0f %10.0f %9.1f%% %s\n",
			"Page File:",
			float64(swapMem.Total)/unitFactor,
			float64(swapMem.Used)/unitFactor,
			float64(swapMem.Free)/unitFactor,
			swapPercent,
			unitLabel,
		)
	}
}
