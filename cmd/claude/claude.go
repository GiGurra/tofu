package claude

import (
	"fmt"
	"os"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/claude/conv"
	claudegit "github.com/gigurra/tofu/cmd/claude/git"
	"github.com/gigurra/tofu/cmd/claude/session"
	"github.com/gigurra/tofu/cmd/claude/setup"
	"github.com/gigurra/tofu/cmd/claude/stats"
	"github.com/gigurra/tofu/cmd/claude/statusbar"
	"github.com/gigurra/tofu/cmd/claude/usage"
	"github.com/gigurra/tofu/cmd/claude/web"
	"github.com/gigurra/tofu/cmd/claude/worktree"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

// Cmd returns the claude subcommand for use in other binaries (e.g. tofu claude ...).
func Cmd() *cobra.Command {
	cmd := boa.CmdT[session.NewParams]{
		Use:         "claude",
		Short:       "Claude Code utilities",
		Long:        "Claude Code utilities.\n\nWhen run without a subcommand, starts a new Claude session in the current directory.",
		ParamEnrich: common.DefaultParamEnricher(),
		SubCmds: []*cobra.Command{
			conv.Cmd(),
			session.Cmd(),
			claudegit.Cmd(),
			worktree.Cmd(),
			stats.Cmd(),
			usage.Cmd(),
			setup.Cmd(),
			statusbar.Cmd(),
			web.Cmd(),
		},
		RunFunc: func(params *session.NewParams, cmd *cobra.Command, args []string) {
			if err := session.RunNew(params); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
	cmd.Args = cobra.ArbitraryArgs
	return cmd
}
