package uwu

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
	Text    []string `pos:"true" optional:"true" help:"Text to uwu-ify. If none provided, reads from stdin."`
	Stutter bool     `short:"s" help:"Add stuttering (e.g., h-hello)." default:"false"`
	Emotes  bool     `short:"e" help:"Add random emotes." default:"false"`
}

var emotes = []string{
	"uwu", "UwU", "owo", "OwO", ">w<", "^w^", ":3", "x3",
	"(ノ◕ヮ◕)ノ*:・゚✧", "(◕ᴗ◕✿)", "(*^ω^)", "(✿◠‿◠)",
}

var replacements = []struct {
	old string
	new string
}{
	{"r", "w"},
	{"l", "w"},
	{"R", "W"},
	{"L", "W"},
	{"no", "nyo"},
	{"No", "Nyo"},
	{"NO", "NYO"},
	{"na", "nya"},
	{"Na", "Nya"},
	{"NA", "NYA"},
	{"ne", "nye"},
	{"Ne", "Nye"},
	{"NE", "NYE"},
	{"ni", "nyi"},
	{"Ni", "Nyi"},
	{"NI", "NYI"},
	{"nu", "nyu"},
	{"Nu", "Nyu"},
	{"NU", "NYU"},
	{"ove", "uv"},
	{"Ove", "Uv"},
	{"OVE", "UV"},
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "uwu",
		Short:       "UwU-ify text",
		Long:        "Transform text into uwu-speak. Pipe your error logs through it for comfort.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	if len(params.Text) > 0 {
		text := strings.Join(params.Text, " ")
		fmt.Println(uwuify(text, params.Stutter, params.Emotes))
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			fmt.Println(uwuify(scanner.Text(), params.Stutter, params.Emotes))
		}
	}
}

func uwuify(text string, stutter bool, addEmotes bool) string {
	result := text

	for _, r := range replacements {
		result = strings.ReplaceAll(result, r.old, r.new)
	}

	if stutter {
		result = addStutter(result)
	}

	if addEmotes && len(result) > 0 {
		result = result + " " + emotes[rand.Intn(len(emotes))]
	}

	return result
}

func addStutter(text string) string {
	words := strings.Fields(text)
	for i, word := range words {
		if len(word) > 0 && rand.Float32() < 0.2 {
			first := strings.ToLower(string(word[0]))
			if first >= "a" && first <= "z" {
				words[i] = first + "-" + word
			}
		}
	}
	return strings.Join(words, " ")
}
