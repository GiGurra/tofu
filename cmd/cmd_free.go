package cmd

import (
	"fmt"
	"os"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/cobra"
)

type FreeParams struct {
	MegaBytes bool `short:"m" help:"Display output in megabytes."`
	GigaBytes bool `short:"g" help:"Display output in gigabytes."`
}

func FreeCmd() *cobra.Command {
	return boa.CmdT[FreeParams]{
		Use:         "free",
		Short:       "Display amount of free and used memory in the system",
		Long: `Display the total, used, and free amount of physical and swap memory in the system.
By default, the output is in kilobytes. Use -m for megabytes or -g for gigabytes.`, 
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *FreeParams, cmd *cobra.Command, args []string) {
			if err := runFree(params); err != nil {
				fmt.Fprintf(os.Stderr, "free: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runFree(params *FreeParams) error {
	virtualMem, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("failed to get virtual memory info: %w", err)
	}

	swapMem, err := mem.SwapMemory()
	if err != nil {
		return fmt.Errorf("failed to get swap memory info: %w", err)
	}

	unitFactor := float64(1)
	unitLabel := ""

	if params.GigaBytes {
		unitFactor = 1024 * 1024 * 1024
		unitLabel = "GiB"
	} else if params.MegaBytes {
		unitFactor = 1024 * 1024
		unitLabel = "MiB"
	} else { // Default to KiB
		unitFactor = 1024
		unitLabel = "KiB"
	}

	fmt.Printf("%12s %10s %10s %10s %10s %10s %10s\n", "", "total", "used", "free", "shared", "buff/cache", "available")
	fmt.Printf("%12s %10.0f %10.0f %10.0f %10.0f %10.0f %10.0f %s\n",
		"Mem:",
		float64(virtualMem.Total)/unitFactor,
		float64(virtualMem.Used)/unitFactor,
		float64(virtualMem.Free)/unitFactor,
		float64(virtualMem.Shared)/unitFactor,
		float64(virtualMem.Buffers+virtualMem.Cached)/unitFactor,
		float64(virtualMem.Available)/unitFactor,
		unitLabel,
	)
	fmt.Printf("%12s %10.0f %10.0f %10.0f %s\n",
		"Swap:",
		float64(swapMem.Total)/unitFactor,
		float64(swapMem.Used)/unitFactor,
		float64(swapMem.Free)/unitFactor,
		unitLabel,
	)

	return nil
}
