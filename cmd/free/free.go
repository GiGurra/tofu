package free

import (
	"fmt"
	"os"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/cobra"
)

type Params struct {
	MegaBytes bool `short:"m" help:"Display output in megabytes."`
	GigaBytes bool `short:"g" help:"Display output in gigabytes."`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:   "free",
		Short: "Display amount of free and used memory in the system",
		Long: `Display the total, used, and free amount of physical and swap memory in the system.
By default, the output is in kilobytes. Use -m for megabytes or -g for gigabytes.`,
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := runFree(params); err != nil {
				fmt.Fprintf(os.Stderr, "free: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runFree(params *Params) error {
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

	printMemoryInfo(os.Stdout, virtualMem, swapMem, unitFactor, unitLabel)

	return nil
}
