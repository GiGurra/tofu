package claude

import (
	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/claude/conv"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	return boa.CmdT[boa.NoParams]{
		Use:   "claude",
		Short: "Claude Code utilities",
		SubCmds: []*cobra.Command{
			conv.Cmd(),
		},
	}.ToCobra()
}
