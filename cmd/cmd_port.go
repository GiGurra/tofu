package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"text/tabwriter"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
)

type PortParams struct {
	PortNum int  `pos:"true" optional:"true" help:"Specific port to inspect."`
	Kill    bool `short:"k" help:"Kill the process listening on the specified port." default:"false"`
	UDP     bool `short:"u" help:"Include UDP ports (TCP is default)." default:"false"`
	All     bool `short:"a" help:"Show all ports (not just listening)." default:"false"`
}

// runPort now takes io.Writer for stdout and stderr for testability
func runPort(params *PortParams, stdout, stderr io.Writer) error {
	if params.Kill && params.PortNum == 0 {
		return fmt.Errorf("--kill requires a specific port number")
	}

	kind := "inet" // TCP + UDP
	if !params.UDP && !params.All {
		kind = "tcp"
	}

	conns, err := net.Connections(kind)
	if err != nil {
		return fmt.Errorf("failed to get connections: %w", err)
	}

	var targets []PortInfo

	for _, conn := range conns {
		if !params.All && conn.Status != "LISTEN" {
			continue
		}

		if params.PortNum != 0 && int(conn.Laddr.Port) != params.PortNum {
			continue
		}

		// If looking for specific port, we match exact.
		// If listing all, we might filter duplicates (same process listening on IPv4 and IPv6)

		proc, err := process.NewProcess(conn.Pid)
		name := "?"
		if err == nil {
			name, _ = proc.Name()
		}

		targets = append(targets, PortInfo{
			Proto:   getProto(conn.Type),
			Port:    int(conn.Laddr.Port),
			Address: conn.Laddr.IP,
			PID:     conn.Pid,
			Name:    name,
			Status:  conn.Status,
		})
	}

	// Deduplicate somewhat? Sockets often appear twice for ipv4/ipv6
	// But sometimes different processes bind same port (SO_REUSEPORT)?
	// Let's sort them.
	sort.Slice(targets, func(i, j int) bool {
		if targets[i].Port != targets[j].Port {
			return targets[i].Port < targets[j].Port
		}
		return targets[i].PID < targets[j].PID
	})

	if len(targets) == 0 {
		if params.PortNum != 0 {
			return fmt.Errorf("no process found on port %d", params.PortNum)
		}
		fmt.Fprintln(stdout, "No matching connections found.")
		return nil
	}

	if params.Kill {
		// Kill logic
		// We might have multiple PIDs for one port (rare but possible, or just multiple entries for same PID)
		pidsKilled := make(map[int32]bool)

		for _, t := range targets {
			if pidsKilled[t.PID] {
				continue
			}

			proc, err := process.NewProcess(t.PID)
			if err != nil {
				fmt.Fprintf(stderr, "Could not find process %d: %v\n", t.PID, err)
				continue
			}

			fmt.Fprintf(stdout, "Killing process %d (%s) on port %d...\n", t.PID, t.Name, t.Port)
			if err := proc.Kill(); err != nil {
				fmt.Fprintf(stderr, "Failed to kill PID %d: %v\n", t.PID, err)
			} else {
				fmt.Fprintf(stdout, "Killed PID %d\n", t.PID)
				pidsKilled[t.PID] = true
			}
		}
		return nil
	}

	// List logic
	w := tabwriter.NewWriter(stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "PROTO\tPORT\tPID\tPROCESS\tSTATUS\tADDRESS")
	for _, t := range targets {
		fmt.Fprintf(w, "%s\t%d\t%d\t%s\t%s\t%s\n", t.Proto, t.Port, t.PID, t.Name, t.Status, t.Address)
	}
	w.Flush()

	return nil
}

type PortInfo struct {
	Proto   string
	Port    int
	Address string
	PID     int32
	Name    string
	Status  string
}

func getProto(t uint32) string {
	// gopsutil constants: 1=TCP, 2=UDP
	switch t {
	case 1:
		return "TCP"
	case 2:
		return "UDP"
	default:
		return strconv.Itoa(int(t))
	}
}

func PortCmd() *cobra.Command {
	return boa.CmdT[PortParams]{
		Use:         "port",
		Short:       "List or kill processes by port",
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *PortParams, cmd *cobra.Command, args []string) {
			if err := runPort(params, os.Stdout, os.Stderr); err != nil {
				fmt.Fprintf(os.Stderr, "port: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}
