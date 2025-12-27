package ping

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

type Params struct {
	Host     string  `pos:"true" required:"true" help:"Host to ping."`
	Count    int     `short:"c" optional:"true" help:"Stop after sending count ECHO_REQUEST packets." default:"0"`
	Interval float64 `short:"i" optional:"true" help:"Wait interval seconds between sending each packet." default:"1"`
	Timeout  float64 `short:"W" optional:"true" help:"Time to wait for a response, in seconds." default:"5"`
	IPv4     bool    `short:"4" optional:"true" help:"Use IPv4 only."`
	IPv6     bool    `short:"6" optional:"true" help:"Use IPv6 only."`
}

type pingStats struct {
	transmitted int
	received    int
	minRTT      time.Duration
	maxRTT      time.Duration
	totalRTT    time.Duration
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "ping",
		Short:       "Send ICMP ECHO_REQUEST to network hosts",
		Long:        "ping uses the ICMP protocol's mandatory ECHO_REQUEST datagram to elicit an ICMP ECHO_RESPONSE from a host or gateway. Requires root/sudo on most systems.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			exitCode := Run(params, os.Stdout, os.Stderr)
			os.Exit(exitCode)
		},
	}.ToCobra()
}

func Run(params *Params, stdout, stderr io.Writer) int {
	// Resolve the host
	addrs, err := net.LookupIP(params.Host)
	if err != nil {
		fmt.Fprintf(stderr, "ping: %s: %v\n", params.Host, err)
		return 1
	}

	var addr net.IP
	for _, a := range addrs {
		if params.IPv4 && a.To4() != nil {
			addr = a
			break
		} else if params.IPv6 && a.To4() == nil {
			addr = a
			break
		} else if !params.IPv4 && !params.IPv6 {
			addr = a
			break
		}
	}

	if addr == nil {
		fmt.Fprintf(stderr, "ping: %s: No suitable address found\n", params.Host)
		return 1
	}

	isIPv6 := addr.To4() == nil

	var network string
	if isIPv6 {
		network = "ip6:ipv6-icmp"
	} else {
		network = "ip4:icmp"
	}

	conn, err := icmp.ListenPacket(network, "")
	if err != nil {
		fmt.Fprintf(stderr, "ping: %v (try running with sudo)\n", err)
		return 1
	}
	defer conn.Close()

	fmt.Fprintf(stdout, "PING %s (%s): 56 data bytes\n", params.Host, addr.String())

	stats := &pingStats{
		minRTT: time.Hour,
	}

	// Handle interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	go func() {
		<-sigChan
		done <- true
	}()

	seq := 0
	ticker := time.NewTicker(time.Duration(params.Interval * float64(time.Second)))
	defer ticker.Stop()

	// Send first ping immediately
	sendPing(conn, addr, seq, isIPv6, params, stdout, stderr, stats)
	seq++
	stats.transmitted++

	for {
		select {
		case <-done:
			printStats(params.Host, stats, stdout)
			if stats.received == 0 {
				return 1
			}
			return 0
		case <-ticker.C:
			if params.Count > 0 && stats.transmitted >= params.Count {
				printStats(params.Host, stats, stdout)
				if stats.received == 0 {
					return 1
				}
				return 0
			}
			sendPing(conn, addr, seq, isIPv6, params, stdout, stderr, stats)
			seq++
			stats.transmitted++
		}
	}
}

func sendPing(conn *icmp.PacketConn, addr net.IP, seq int, isIPv6 bool, params *Params, stdout, stderr io.Writer, stats *pingStats) {
	var msgType icmp.Type
	if isIPv6 {
		msgType = ipv6.ICMPTypeEchoRequest
	} else {
		msgType = ipv4.ICMPTypeEcho
	}

	msg := icmp.Message{
		Type: msgType,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  seq,
			Data: make([]byte, 56),
		},
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		fmt.Fprintf(stderr, "ping: %v\n", err)
		return
	}

	dst := &net.IPAddr{IP: addr}

	start := time.Now()
	if _, err := conn.WriteTo(msgBytes, dst); err != nil {
		fmt.Fprintf(stderr, "ping: %v\n", err)
		return
	}

	conn.SetReadDeadline(time.Now().Add(time.Duration(params.Timeout * float64(time.Second))))

	reply := make([]byte, 1500)
	n, _, err := conn.ReadFrom(reply)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			fmt.Fprintf(stdout, "Request timeout for icmp_seq %d\n", seq)
			return
		}
		fmt.Fprintf(stderr, "ping: %v\n", err)
		return
	}

	rtt := time.Since(start)

	var proto int
	if isIPv6 {
		proto = 58 // ICMPv6
	} else {
		proto = 1 // ICMP
	}

	parsedMsg, err := icmp.ParseMessage(proto, reply[:n])
	if err != nil {
		fmt.Fprintf(stderr, "ping: %v\n", err)
		return
	}

	switch parsedMsg.Type {
	case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeEchoReply:
		stats.received++
		stats.totalRTT += rtt
		if rtt < stats.minRTT {
			stats.minRTT = rtt
		}
		if rtt > stats.maxRTT {
			stats.maxRTT = rtt
		}
		fmt.Fprintf(stdout, "64 bytes from %s: icmp_seq=%d time=%.3f ms\n",
			addr.String(), seq, float64(rtt.Microseconds())/1000.0)
	default:
		fmt.Fprintf(stdout, "Unexpected ICMP message type: %v\n", parsedMsg.Type)
	}
}

func printStats(host string, stats *pingStats, stdout io.Writer) {
	fmt.Fprintf(stdout, "\n--- %s ping statistics ---\n", host)
	loss := float64(stats.transmitted-stats.received) / float64(stats.transmitted) * 100
	fmt.Fprintf(stdout, "%d packets transmitted, %d packets received, %.1f%% packet loss\n",
		stats.transmitted, stats.received, loss)

	if stats.received > 0 {
		avg := stats.totalRTT / time.Duration(stats.received)
		fmt.Fprintf(stdout, "round-trip min/avg/max = %.3f/%.3f/%.3f ms\n",
			float64(stats.minRTT.Microseconds())/1000.0,
			float64(avg.Microseconds())/1000.0,
			float64(stats.maxRTT.Microseconds())/1000.0)
	}
}
