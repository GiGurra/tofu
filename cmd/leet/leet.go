package leet

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
	Text   []string `pos:"true" optional:"true" help:"Text to convert to l33tsp34k. If none provided, reads from stdin."`
	Level  int      `short:"l" help:"Leetness level: 1=basic, 2=intermediate, 3=advanced." default:"2"`
	Random bool     `short:"r" help:"Randomly vary substitutions for more authentic look." default:"false"`
}

// Basic substitutions (level 1)
var basicSubs = map[rune]string{
	'a': "4", 'A': "4",
	'e': "3", 'E': "3",
	'i': "1", 'I': "1",
	'o': "0", 'O': "0",
}

// Intermediate substitutions (level 2) - adds more
var intermediateSubs = map[rune]string{
	'a': "4", 'A': "4",
	'e': "3", 'E': "3",
	'i': "1", 'I': "1",
	'o': "0", 'O': "0",
	's': "5", 'S': "5",
	't': "7", 'T': "7",
}

// Advanced substitutions (level 3) - full leet
var advancedSubs = map[rune]string{
	'a': "4", 'A': "4",
	'b': "8", 'B': "8",
	'e': "3", 'E': "3",
	'g': "9", 'G': "9",
	'i': "1", 'I': "1",
	'l': "1", 'L': "1",
	'o': "0", 'O': "0",
	's': "5", 'S': "5",
	't': "7", 'T': "7",
	'z': "2", 'Z': "2",
}

// Alternative substitutions for random mode
var altSubs = map[rune][]string{
	'a': {"4", "@", "/-\\"},
	'A': {"4", "@", "/-\\"},
	'b': {"8", "|3"},
	'B': {"8", "|3"},
	'e': {"3"},
	'E': {"3"},
	'g': {"9", "6"},
	'G': {"9", "6"},
	'i': {"1", "!", "|"},
	'I': {"1", "!", "|"},
	'l': {"1", "|"},
	'L': {"1", "|"},
	'o': {"0"},
	'O': {"0"},
	's': {"5", "$"},
	'S': {"5", "$"},
	't': {"7", "+"},
	'T': {"7", "+"},
	'z': {"2"},
	'Z': {"2"},
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "leet",
		Short:       "Convert text to l33tsp34k",
		Long:        "7r4n5f0rm 73x7 1n70 l337sp34k. Perfect for feeling like a 90s hacker.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	if len(params.Text) > 0 {
		text := strings.Join(params.Text, " ")
		fmt.Println(leetify(text, params.Level, params.Random))
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			fmt.Println(leetify(scanner.Text(), params.Level, params.Random))
		}
	}
}

func leetify(text string, level int, random bool) string {
	var subs map[rune]string
	switch level {
	case 1:
		subs = basicSubs
	case 3:
		subs = advancedSubs
	default:
		subs = intermediateSubs
	}

	var result strings.Builder
	for _, r := range text {
		if random {
			if alts, ok := altSubs[r]; ok {
				// Only substitute sometimes for more natural look
				if rand.Float32() < 0.7 {
					result.WriteString(alts[rand.Intn(len(alts))])
				} else {
					result.WriteRune(r)
				}
				continue
			}
		} else {
			if sub, ok := subs[r]; ok {
				result.WriteString(sub)
				continue
			}
		}
		result.WriteRune(r)
	}
	return result.String()
}
