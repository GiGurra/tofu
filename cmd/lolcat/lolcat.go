package lolcat

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Text   []string `pos:"true" optional:"true" help:"Text to colorize. If none provided, reads from stdin."`
	Freq   float64  `short:"f" help:"Rainbow frequency." default:"0.1"`
	Spread float64  `short:"p" help:"Rainbow spread." default:"3.0"`
	Seed   float64  `short:"s" help:"Rainbow seed (starting point)." default:"0"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "lolcat",
		Short:       "Rainbow colorize text",
		Long:        "Output text in rainbow colors. Makes everything better.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	lineNum := 0

	if len(params.Text) > 0 {
		text := strings.Join(params.Text, " ")
		fmt.Println(rainbowLine(text, lineNum, params.Freq, params.Spread, params.Seed))
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			fmt.Println(rainbowLine(scanner.Text(), lineNum, params.Freq, params.Spread, params.Seed))
			lineNum++
		}
	}
}

func rainbowLine(text string, lineNum int, freq, spread, seed float64) string {
	var result strings.Builder

	for i, r := range text {
		if r == ' ' || r == '\t' {
			result.WriteRune(r)
			continue
		}

		// Calculate rainbow position
		pos := seed + float64(lineNum)/spread + float64(i)/spread
		color := rainbow(freq, pos)

		// Write ANSI color code + character + reset
		result.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm%c\x1b[0m", color.r, color.g, color.b, r))
	}

	return result.String()
}

type rgb struct {
	r, g, b uint8
}

func rainbow(freq, pos float64) rgb {
	return rgb{
		r: uint8(math.Sin(freq*pos+0)*127 + 128),
		g: uint8(math.Sin(freq*pos+2*math.Pi/3)*127 + 128),
		b: uint8(math.Sin(freq*pos+4*math.Pi/3)*127 + 128),
	}
}
