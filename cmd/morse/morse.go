package morse

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
	Text   []string `pos:"true" optional:"true" help:"Text to encode/decode. If none provided, reads from stdin."`
	Decode bool     `short:"d" help:"Decode morse code to text." default:"false"`
	Beep   bool     `short:"b" help:"Play audio beeps while encoding (requires CGO on Linux)." default:"false"`
	WPM    int      `short:"w" help:"Words per minute for audio playback." default:"15"`
}

var toMorse = map[rune]string{
	'A': ".-", 'B': "-...", 'C': "-.-.", 'D': "-..", 'E': ".",
	'F': "..-.", 'G': "--.", 'H': "....", 'I': "..", 'J': ".---",
	'K': "-.-", 'L': ".-..", 'M': "--", 'N': "-.", 'O': "---",
	'P': ".--.", 'Q': "--.-", 'R': ".-.", 'S': "...", 'T': "-",
	'U': "..-", 'V': "...-", 'W': ".--", 'X': "-..-", 'Y': "-.--",
	'Z': "--..",
	'0': "-----", '1': ".----", '2': "..---", '3': "...--", '4': "....-",
	'5': ".....", '6': "-....", '7': "--...", '8': "---..", '9': "----.",
	'.': ".-.-.-", ',': "--..--", '?': "..--..", '\'': ".----.",
	'!': "-.-.--", '/': "-..-.", '(': "-.--.", ')': "-.--.-",
	'&': ".-...", ':': "---...", ';': "-.-.-.", '=': "-...-",
	'+': ".-.-.", '-': "-....-", '_': "..--.-", '"': ".-..-.",
	'$': "...-..-", '@': ".--.-.",
	' ': "/",
}

var fromMorse map[string]rune

func init() {
	fromMorse = make(map[string]rune)
	for k, v := range toMorse {
		if k != ' ' {
			fromMorse[v] = k
		}
	}
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "morse",
		Short:       "Encode/decode Morse code",
		Long:        "Convert text to Morse code or decode Morse code back to text. Use -b for audio beeps.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	if len(params.Text) > 0 {
		text := strings.Join(params.Text, " ")
		if params.Decode {
			fmt.Println(decode(text))
		} else {
			encoded := encode(text)
			fmt.Println(encoded)
			if params.Beep {
				playMorse(encoded, params.WPM)
			}
		}
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if params.Decode {
				fmt.Println(decode(scanner.Text()))
			} else {
				encoded := encode(scanner.Text())
				fmt.Println(encoded)
				if params.Beep {
					playMorse(encoded, params.WPM)
				}
			}
		}
	}
}

func encode(text string) string {
	var result []string
	for _, r := range strings.ToUpper(text) {
		if code, ok := toMorse[r]; ok {
			result = append(result, code)
		}
	}
	return strings.Join(result, " ")
}

func decode(morse string) string {
	var result strings.Builder
	words := strings.Split(morse, " / ")
	for i, word := range words {
		if i > 0 {
			result.WriteRune(' ')
		}
		codes := strings.Fields(word)
		for _, code := range codes {
			if r, ok := fromMorse[code]; ok {
				result.WriteRune(r)
			}
		}
	}
	return result.String()
}
