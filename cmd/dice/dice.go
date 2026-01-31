package dice

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Dice []string `pos:"true" optional:"true" help:"Dice notation (e.g., 2d20+5, d6, 3d8-2). Default: 1d6" default:"1d6"`
}

// dicePattern matches dice notation like "2d20+5", "d6", "3d8-2"
var dicePattern = regexp.MustCompile(`^(\d*)d(\d+)([+-]\d+)?$`)

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "dice",
		Short:       "Roll dice using standard notation",
		Long:        "Roll dice using D&D-style notation. Examples: d20, 2d6, 3d8+5, 1d20-2",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := Run(params); err != nil {
				fmt.Fprintf(os.Stderr, "dice: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func Run(params *Params) error {
	for _, notation := range params.Dice {
		result, breakdown, err := Roll(notation)
		if err != nil {
			return err
		}
		if breakdown != "" {
			fmt.Printf("%s: %d (%s)\n", notation, result, breakdown)
		} else {
			fmt.Printf("%s: %d\n", notation, result)
		}
	}
	return nil
}

func Roll(notation string) (int, string, error) {
	notation = strings.ToLower(strings.TrimSpace(notation))

	matches := dicePattern.FindStringSubmatch(notation)
	if matches == nil {
		return 0, "", fmt.Errorf("invalid dice notation: %s (use format like 2d20+5)", notation)
	}

	// Parse number of dice (default 1)
	numDice := 1
	if matches[1] != "" {
		var err error
		numDice, err = strconv.Atoi(matches[1])
		if err != nil {
			return 0, "", fmt.Errorf("invalid number of dice: %s", matches[1])
		}
	}

	// Parse die sides
	sides, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, "", fmt.Errorf("invalid die sides: %s", matches[2])
	}
	if sides < 1 {
		return 0, "", fmt.Errorf("die must have at least 1 side")
	}

	// Parse modifier (default 0)
	modifier := 0
	if matches[3] != "" {
		modifier, err = strconv.Atoi(matches[3])
		if err != nil {
			return 0, "", fmt.Errorf("invalid modifier: %s", matches[3])
		}
	}

	// Roll the dice
	rolls := make([]int, numDice)
	total := 0
	for i := 0; i < numDice; i++ {
		rolls[i] = rand.Intn(sides) + 1
		total += rolls[i]
	}
	total += modifier

	// Build breakdown string
	var breakdown string
	if numDice > 1 || modifier != 0 {
		parts := make([]string, len(rolls))
		for i, r := range rolls {
			parts[i] = strconv.Itoa(r)
		}
		breakdown = strings.Join(parts, "+")
		if modifier > 0 {
			breakdown += fmt.Sprintf("+%d", modifier)
		} else if modifier < 0 {
			breakdown += fmt.Sprintf("%d", modifier)
		}
	}

	return total, breakdown, nil
}
