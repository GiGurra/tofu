package cmd

import (
	"fmt"
	"os"
	"os/user"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
)

type PsParams struct {
	Full       bool     `short:"f" help:"Display full format listing."`
	Users      []string `short:"u" optional:"true" help:"Filter by username(s)."`
	Pids       []int32  `short:"p" optional:"true" help:"Filter by PID(s)."`
	Name       string   `short:"n" optional:"true" help:"Filter by command name (substring)."`
	Current    bool     `short:"c" help:"Show only processes owned by the current user."`
	Invert     bool     `short:"v" help:"Invert filtering (matches non-matching processes)."`
	NoTruncate bool     `short:"N" help:"Do not truncate command line output."`
}

func PsCmd() *cobra.Command {
	return boa.CmdT[PsParams]{
		Use:   "ps",
		Short: "Report a snapshot of the current processes",
		Long: `Displays information about a selection of the active processes.
By default, it lists all processes with a minimal set of columns.
Use -f for a full format listing.
Filters can be combined (AND logic). Use -v to invert the filter.
Use -N to prevent truncation of command line output.`,
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *PsParams, cmd *cobra.Command, args []string) {
			if err := runPs(params); err != nil {
				fmt.Fprintf(os.Stderr, "ps: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runPs(params *PsParams) error {
	procs, err := process.Processes()
	if err != nil {
		return fmt.Errorf("failed to list processes: %w", err)
	}

	currentUser, err := user.Current()
	currentUsername := ""
	if err == nil {
		currentUsername = currentUser.Username
		// On Windows, username might be qualified with domain (DOMAIN\User)
		// gopsutil might return just User or DOMAIN\User.
		// For robustness, we might need flexible matching, but exact match is a good start.
	} else if params.Current {
		return fmt.Errorf("failed to determine current user: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	// Header
	if params.Full {
		fmt.Fprintln(w, "PID\tPPID\tUSER\tSTATUS\t%CPU\t%MEM\tCOMMAND")
	} else {
		fmt.Fprintln(w, "PID\tCOMMAND")
	}

	// Sort by PID
	sort.Slice(procs, func(i, j int) bool {
		return procs[i].Pid < procs[j].Pid
	})

	for _, p := range procs {
		if !shouldInclude(p, params, currentUsername) {
			continue
		}

		pid := p.Pid
		name, _ := p.Name()
		if name == "" {
			name = "[unknown]"
		}

		if params.Full {
			ppid, _ := p.Ppid()
			username, _ := p.Username()
			if username == "" {
				username = "-"
			}

			status, _ := p.Status()
			cpu, _ := p.CPUPercent()
			mem, _ := p.MemoryPercent()
			cmdline, _ := p.Cmdline()
			if cmdline == "" {
				cmdline = name
			} else if !params.NoTruncate && len(cmdline) > 50 {
				cmdline = cmdline[:47] + "..."
			}

			statusStr := ""
			if len(status) > 0 {
				statusStr = status[0]
			}

			fmt.Fprintf(w, "%d\t%d\t%s\t%s\t%.1f\t%.1f\t%s\n",
				pid, ppid, username, statusStr, cpu, mem, cmdline)
		} else {
			fmt.Fprintf(w, "%d\t%s\n", pid, name)
		}
	}

	return nil
}

func shouldInclude(p *process.Process, params *PsParams, currentUsername string) bool {
	// If no filters are active, include everything
	if len(params.Users) == 0 && len(params.Pids) == 0 && params.Name == "" && !params.Current {
		return true
	}

	matched := true

	// PID Filter
	if len(params.Pids) > 0 {
		pidMatch := false
		for _, pid := range params.Pids {
			if p.Pid == pid {
				pidMatch = true
				break
			}
		}
		if !pidMatch {
			matched = false
		}
	}

	// Name Filter (if still matched)
	if matched && params.Name != "" {
		name, _ := p.Name()
		if !strings.Contains(name, params.Name) {
			matched = false
		}
	}

	// User Filter (if still matched)
	if matched && len(params.Users) > 0 {
		userMatch := false
		username, _ := p.Username()
		for _, u := range params.Users {
			if username == u {
				userMatch = true
				break
			}
		}
		if !userMatch {
			matched = false
		}
	}

	// Current User Filter (if still matched)
	if matched && params.Current {
		username, _ := p.Username()
		// Simple string comparison.
		// Note: On Windows, username might differ slightly (case, domain prefix).
		// Ideally we'd compare UIDs/SIDs, but gopsutil Uids() returns []int32 on unix and error on windows sometimes?
		// Stick to string for now.
		if username != currentUsername {
			matched = false
		}
	}

	if params.Invert {
		return !matched
	}
	return matched
}
