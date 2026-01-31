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
	JSON  bool   `long:"json" help:"Output as JSON"`
	All   bool   `short:"a" long:"all" help:"Include exited sessions"`
	Watch bool   `short:"w" long:"watch" help:"Interactive watch mode with auto-refresh"`
	Sort  string `short:"s" long:"sort" optional:"true" help:"Sort by column" alts:"id,directory,status,age,updated"`
	Asc   bool   `long:"asc" help:"Sort ascending (default for id/directory/status)"`
	Desc  bool   `long:"desc" help:"Sort descending (default for updated)"`
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
	// Parse sort options
	sortState := parseSortParams(params.Sort, params.Asc, params.Desc)

	// Interactive watch mode
	if params.Watch {
		return RunWatchMode(params.All, sortState)
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

	// Apply sorting
	SortSessions(filtered, sortState)

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

	t.AppendHeader(table.Row{"ID", "Directory", "Status", "Age", "Updated"})

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
			colorFunc(FormatDuration(time.Since(state.Created))),
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

func parseSortParams(sortBy string, asc, desc bool) SortState {
	var col SortColumn
	switch sortBy {
	case "id":
		col = SortID
	case "directory", "dir":
		col = SortDirectory
	case "status":
		col = SortStatus
	case "age", "created":
		col = SortAge
	case "updated", "time":
		col = SortUpdated
	default:
		col = SortNone
	}

	// Default direction depends on column
	// Time-based columns default to descending (most recent first)
	// Others default to ascending
	var dir SortDirection
	if asc {
		dir = SortAsc
	} else if desc {
		dir = SortDesc
	} else if col == SortUpdated || col == SortAge {
		dir = SortDesc // Smart default for time
	} else {
		dir = SortAsc
	}

	return SortState{Column: col, Direction: dir}
}
