package blame

import (
	"fmt"
	"math/rand"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

var culprits = []string{
	"cosmic rays",
	"the intern",
	"the previous developer",
	"mercury retrograde",
	"sunspots",
	"a rogue cosmic bit flip",
	"the heat death of the universe",
	"quantum fluctuations",
	"the cloud",
	"microservices",
	"kubernetes",
	"docker",
	"the network",
	"DNS",
	"the database",
	"the cache",
	"the load balancer",
	"the firewall",
	"the vendor",
	"the API",
	"JavaScript",
	"NPM",
	"left-pad",
	"node_modules",
	"the monolith",
	"technical debt",
	"the sprint deadline",
	"the product manager",
	"the requirements",
	"the spec",
	"JIRA",
	"Confluence",
	"the meeting that could have been an email",
	"agile",
	"the retrospective action items",
	"the deployment pipeline",
	"the staging environment",
	"the test data",
	"the mocks",
	"the stubs",
	"legacy code",
	"that one guy who left 3 years ago",
	"Stack Overflow",
	"ChatGPT",
	"the AI",
	"the linter",
	"the formatter",
	"tabs vs spaces",
	"the merge conflict",
	"git rebase",
	"a mass coronal ejection event",
	"the butterfly effect",
	"a developer who mass-quit Vim",
	"someone who mass-quit Vim",
	"undefined behavior",
	"a null pointer",
	"off-by-one errors",
}

type Params struct {
	Count int `short:"n" help:"Number of things to blame." default:"1"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "blame",
		Short:       "Randomly blame something for the bug",
		Long:        "When things go wrong, this command will tell you what to blame.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			for i := 0; i < params.Count; i++ {
				culprit := culprits[rand.Intn(len(culprits))]
				fmt.Printf("It's clearly %s.\n", culprit)
			}
		},
	}.ToCobra()
}
