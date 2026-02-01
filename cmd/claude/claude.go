package claude

import (
	"fmt"
	"os"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/claude/conv"
	"github.com/gigurra/tofu/cmd/claude/git"
	"github.com/gigurra/tofu/cmd/claude/session"
	"github.com/gigurra/tofu/cmd/claude/setup"
	"github.com/gigurra/tofu/cmd/claude/usage"
	"github.com/gigurra/tofu/cmd/claude/worktree"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	return boa.CmdT[session.NewParams]{
		Use:         "claude",
		Short:       "Claude Code utilities",
		Long:        "Claude Code utilities.\n\nWhen run without a subcommand, starts a new Claude session in the current directory.",
		ParamEnrich: common.DefaultParamEnricher(),
		SubCmds: []*cobra.Command{
			conv.Cmd(),
			session.Cmd(),
			git.Cmd(),
			worktree.Cmd(),
			usage.Cmd(),
			setup.Cmd(),
		},
		RunFunc: func(params *session.NewParams, cmd *cobra.Command, args []string) {
			// Default to starting a new session
			if err := session.RunNew(params); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}
