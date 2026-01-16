package pick

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Items []string `pos:"true" optional:"true" help:"Items to pick from. If none provided, reads from stdin."`
	Count int      `short:"n" help:"Number of items to pick." default:"1"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:     "pick",
		Short:   "Randomly pick from a list",
		Long:        "Randomly select items from arguments or stdin. Great for settling debates or choosing lunch spots.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := Run(params); err != nil {
				fmt.Fprintf(os.Stderr, "pick: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func Run(params *Params) error {
	items := params.Items

	// If no args, read from stdin
	if len(items) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				items = append(items, line)
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
	}

	if len(items) == 0 {
		return fmt.Errorf("no items to pick from")
	}

	if params.Count > len(items) {
		return fmt.Errorf("cannot pick %d items from %d options", params.Count, len(items))
	}

	// Shuffle and pick
	shuffled := make([]string, len(items))
	copy(shuffled, items)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	for i := 0; i < params.Count; i++ {
		fmt.Println(shuffled[i])
	}

	return nil
}
