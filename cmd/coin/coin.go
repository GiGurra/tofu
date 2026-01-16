package coin

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Count    int  `short:"n" help:"Number of flips." default:"1"`
	Animate  bool `short:"a" help:"Show flip animation." default:"false"`
}

var frames = []string{"🪙", "⚪", "🪙", "⚫"}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "coin",
		Short:       "Flip a coin",
		Long:        "Flip a coin to make important decisions. Supports multiple flips and optional animation.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	heads := 0
	tails := 0

	for i := 0; i < params.Count; i++ {
		if params.Animate && params.Count == 1 {
			animate()
		}

		if rand.Intn(2) == 0 {
			heads++
			if params.Count == 1 {
				fmt.Println("Heads!")
			}
		} else {
			tails++
			if params.Count == 1 {
				fmt.Println("Tails!")
			}
		}
	}

	if params.Count > 1 {
		fmt.Printf("Heads: %d, Tails: %d\n", heads, tails)
	}
}

func animate() {
	for i := 0; i < 8; i++ {
		fmt.Printf("\r%s Flipping...  ", frames[i%len(frames)])
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Print("\r")
}
