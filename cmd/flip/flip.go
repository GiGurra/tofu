package flip

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Text  []string `pos:"true" optional:"true" help:"Text to flip. If none provided, reads from stdin."`
	Table bool     `short:"t" help:"Add table flip emote." default:"false"`
}

var flipMap = map[rune]rune{
	'a': 'ɐ', 'b': 'q', 'c': 'ɔ', 'd': 'p', 'e': 'ǝ', 'f': 'ɟ', 'g': 'ƃ',
	'h': 'ɥ', 'i': 'ᴉ', 'j': 'ɾ', 'k': 'ʞ', 'l': 'l', 'm': 'ɯ', 'n': 'u',
	'o': 'o', 'p': 'd', 'q': 'b', 'r': 'ɹ', 's': 's', 't': 'ʇ', 'u': 'n',
	'v': 'ʌ', 'w': 'ʍ', 'x': 'x', 'y': 'ʎ', 'z': 'z',
	'A': '∀', 'B': 'q', 'C': 'Ɔ', 'D': 'p', 'E': 'Ǝ', 'F': 'Ⅎ', 'G': '⅁',
	'H': 'H', 'I': 'I', 'J': 'ſ', 'K': 'ʞ', 'L': '˥', 'M': 'W', 'N': 'N',
	'O': 'O', 'P': 'Ԁ', 'Q': 'Q', 'R': 'ɹ', 'S': 'S', 'T': '⊥', 'U': '∩',
	'V': 'Λ', 'W': 'M', 'X': 'X', 'Y': '⅄', 'Z': 'Z',
	'0': '0', '1': 'Ɩ', '2': 'ᄅ', '3': 'Ɛ', '4': 'ㄣ', '5': 'ϛ', '6': '9',
	'7': 'ㄥ', '8': '8', '9': '6',
	'.': '˙', ',': '\'', '\'': ',', '"': '„', '`': ',', '?': '¿', '!': '¡',
	'[': ']', ']': '[', '(': ')', ')': '(', '{': '}', '}': '{',
	'<': '>', '>': '<', '&': '⅋', '_': '‾', ';': '؛', '∴': '∵',
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "flip",
		Short:       "Flip text upside down",
		Long:        "Transform text to appear upside down. (╯°□°)╯︵ ┻━┻",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	if len(params.Text) > 0 {
		text := strings.Join(params.Text, " ")
		output := flipText(text)
		if params.Table {
			fmt.Printf("(╯°□°)╯︵ %s\n", output)
		} else {
			fmt.Println(output)
		}
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			output := flipText(scanner.Text())
			if params.Table {
				fmt.Printf("(╯°□°)╯︵ %s\n", output)
			} else {
				fmt.Println(output)
			}
		}
	}
}

func flipText(text string) string {
	runes := []rune(text)
	result := make([]rune, len(runes))

	for i, r := range runes {
		if flipped, ok := flipMap[r]; ok {
			result[len(runes)-1-i] = flipped
		} else {
			result[len(runes)-1-i] = r
		}
	}

	return string(result)
}
