package git

import (
	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/git/sync"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	return boa.CmdT[boa.NoParams]{
		Use:   "git",
		Short: "Git utilities",
		SubCmds: []*cobra.Command{
			sync.Cmd(),
		},
	}.ToCobra()
}
