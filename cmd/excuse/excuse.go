package excuse

import (
	"fmt"
	"math/rand"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

var excuses = []string{
	"It works on my machine.",
	"That's not a bug, it's a feature.",
	"It must be a caching issue.",
	"Have you tried clearing your cache?",
	"It worked yesterday.",
	"Someone must have changed something.",
	"That's a known issue.",
	"It's on my list.",
	"The tests passed locally.",
	"It must be a race condition.",
	"DNS.",
	"It's probably cosmic rays.",
	"The documentation is wrong.",
	"It's a dependency issue.",
	"The requirements were unclear.",
	"It's in the backlog.",
	"That's outside the scope.",
	"We'll fix it in v2.",
	"It's a legacy system.",
	"The intern did it.",
	"Git blame says it wasn't me.",
	"The API changed.",
	"It's a timezone issue.",
	"I didn't get the memo.",
	"The server was restarting.",
	"It's an edge case.",
	"The CI/CD pipeline is flaky.",
	"Have you tried turning it off and on again?",
	"It's a network issue.",
	"The database was slow.",
	"It's technically correct.",
	"That's undefined behavior.",
	"It was working before the merge.",
	"The logs don't show anything.",
	"It's probably a memory leak somewhere.",
	"The config is wrong in production.",
	"We're waiting on another team.",
	"That's a browser bug.",
	"It only happens under load.",
	"The firewall is blocking it.",
	"Someone deployed without telling me.",
	"It's deprecated anyway.",
	"The vendor SDK is broken.",
}

type Params struct {
	Count int `short:"n" help:"Number of excuses to generate." default:"1"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "excuse",
		Short:       "Generate programmer excuses",
		Long:        "Generate random programmer excuses for when things go wrong.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			for i := 0; i < params.Count; i++ {
				fmt.Println(excuses[rand.Intn(len(excuses))])
			}
		},
	}.ToCobra()
}
