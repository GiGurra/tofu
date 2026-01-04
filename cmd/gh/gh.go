package gh

import (
	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/gh/listrepos"
	"github.com/gigurra/tofu/cmd/gh/open"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	return boa.CmdT[boa.NoParams]{
		Use:   "gh",
		Short: "GitHub utilities",
		SubCmds: []*cobra.Command{
			listrepos.Cmd(),
			open.Cmd(),
		},
	}.ToCobra()
}
