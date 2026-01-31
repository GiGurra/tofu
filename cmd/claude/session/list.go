package session

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type ListParams struct {
	JSON  bool `long:"json" help:"Output as JSON"`
	All   bool `short:"a" long:"all" help:"Include exited sessions"`
	Watch bool `short:"w" long:"watch" help:"Interactive watch mode with auto-refresh"`
}

func ListCmd() *cobra.Command {
	return boa.CmdT[ListParams]{
		Use:         "ls",
		Aliases:     []string{"list", "status"},
		Short:       "List Claude Code sessions",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *ListParams, cmd *cobra.Command, args []string) {
			if err := runList(params); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runList(params *ListParams) error {
	// Interactive watch mode
	if params.Watch {
		return RunWatchMode(params.All)
	}

	states, err := ListSessionStates()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(states) == 0 {
		fmt.Println("No sessions found")
		fmt.Println("\nStart a new session with: tofu claude session new")
		return nil
	}

	// Refresh status and filter
	var filtered []*SessionState
	for _, state := range states {
		RefreshSessionStatus(state)
		if !params.All && state.Status == StatusExited {
			continue
		}
		filtered = append(filtered, state)
	}

	if len(filtered) == 0 {
		fmt.Println("No active sessions found")
		fmt.Println("\nStart a new session with: tofu claude session new")
		return nil
	}

	if params.JSON {
		data, err := json.MarshalIndent(filtered, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	// Render table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.SetAllowedRowLength(getTermWidth())

	t.AppendHeader(table.Row{"ID", "Directory", "Status", "Updated"})

	for _, state := range filtered {
		colorFunc := getStatusColorFunc(state.Status)
		status := state.Status
		if state.StatusDetail != "" {
			status = status + ": " + state.StatusDetail
		}

		t.AppendRow(table.Row{
			colorFunc(state.ID),
			colorFunc(shortenPathForTable(state.Cwd, 40)),
			colorFunc(status),
			colorFunc(FormatDuration(time.Since(state.Updated))),
		})
	}

	t.Render()

	// Show hint for attaching
	fmt.Printf("\nAttach with: tofu claude session attach <id>\n")

	return nil
}

func getTermWidth() int {
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		return width
	}
	if width, _, err := term.GetSize(int(os.Stderr.Fd())); err == nil && width > 0 {
		return width
	}
	return 120
}

func shortenPathForTable(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "…" + path[len(path)-maxLen+1:]
}

func getStatusColorFunc(status string) func(a ...interface{}) string {
	switch status {
	case StatusIdle:
		return text.FgYellow.Sprint
	case StatusWorking:
		return text.FgGreen.Sprint
	case StatusAwaitingPermission, StatusAwaitingInput:
		return text.FgHiRed.Sprint
	case StatusExited:
		return text.FgHiBlack.Sprint
	default:
		return text.FgHiRed.Sprint // Unknown status = needs attention
	}
}

