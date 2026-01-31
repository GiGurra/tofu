package magic8

import (
	"fmt"
	"math/rand"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

var responses = []string{
	// Positive
	"It is certain.",
	"It is decidedly so.",
	"Without a doubt.",
	"Yes, definitely.",
	"You may rely on it.",
	"As I see it, yes.",
	"Most likely.",
	"Outlook good.",
	"Yes.",
	"Signs point to yes.",
	// Neutral
	"Reply hazy, try again.",
	"Ask again later.",
	"Better not tell you now.",
	"Cannot predict now.",
	"Concentrate and ask again.",
	// Negative
	"Don't count on it.",
	"My reply is no.",
	"My sources say no.",
	"Outlook not so good.",
	"Very doubtful.",
	// Tech-flavored extras
	"Have you tried turning it off and on again?",
	"Works on my machine.",
	"That's a feature, not a bug.",
	"Ship it.",
	"LGTM.",
	"Needs more unit tests.",
	"Ask the senior dev.",
	"Check Stack Overflow.",
	"It depends.",
	"That's out of scope.",
}

type Params struct {
	Question string `pos:"true" optional:"true" help:"Your question (optional, just for fun)."`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "magic8",
		Short:       "Ask the Magic 8-Ball",
		Long:        "Ask the Magic 8-Ball for guidance on important architectural decisions.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	fmt.Printf("🎱 %s\n", responses[rand.Intn(len(responses))])
}
