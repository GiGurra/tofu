package cmd

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

type NcParams struct {
	Args    []string `pos:"true" help:"Host and port (e.g. 'localhost 8080') or just port if listening."`
	Listen  bool     `short:"l" optional:"true" help:"Listen mode, for inbound connects."`
	UDP     bool     `short:"u" optional:"true" help:"Use UDP instead of default TCP."`
	Verbose bool     `short:"v" optional:"true" help:"Verbose mode."`
}

func NcCmd() *cobra.Command {
	return boa.CmdT[NcParams]{
		Use:         "nc",
		Short:       "Netcat clone: Connect to or listen on sockets (TCP/UDP)",
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *NcParams, cmd *cobra.Command, args []string) {
			if len(params.Args) < 1 {
				_ = cmd.Usage()
				os.Exit(1)
			}
			if err := runNc(params, os.Stdin, os.Stdout, os.Stderr); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runNc(params *NcParams, stdin io.Reader, stdout, stderr io.Writer) error {
	host, port, err := parseNcArgs(params.Args, params.Listen)
	if err != nil {
		return err
	}

	protocol := "tcp"
	if params.UDP {
		protocol = "udp"
	}

	address := net.JoinHostPort(host, port)

	if params.Listen {
		return runNcServer(protocol, address, params.Verbose, stdin, stdout, stderr)
	}
	return runNcClient(protocol, address, params.Verbose, stdin, stdout, stderr)
}

func parseNcArgs(args []string, listen bool) (string, string, error) {
	if len(args) == 2 {
		return args[0], args[1], nil
	}
	if len(args) == 1 {
		// If only one arg, it could be a port (for listening) or host:port
		if strings.Contains(args[0], ":") {
			host, port, err := net.SplitHostPort(args[0])
			if err != nil {
				return "", "", fmt.Errorf("invalid address format: %w", err)
			}
			return host, port, nil
		}
		// Assume it's a port if it's a number
		if _, err := strconv.Atoi(args[0]); err == nil {
			if listen {
				return "", args[0], nil // Listen on all interfaces
			}
			return "127.0.0.1", args[0], nil // Default to localhost for client
		}
		return "", "", fmt.Errorf("missing port")
	}
	return "", "", fmt.Errorf("invalid arguments")
}

func runNcClient(protocol, address string, verbose bool, stdin io.Reader, stdout, stderr io.Writer) error {
	if verbose {
		fmt.Fprintf(stderr, "Connecting to %s (%s)...\n", address, protocol)
	}

	conn, err := net.Dial(protocol, address)
	if err != nil {
		return err
	}
	defer conn.Close()

	if verbose {
		fmt.Fprintf(stderr, "Connected!\n")
	}

	return pipeStream(conn, stdin, stdout)
}

func runNcServer(protocol, address string, verbose bool, stdin io.Reader, stdout, stderr io.Writer) error {
	if verbose {
		fmt.Fprintf(stderr, "Listening on %s (%s)...\n", address, protocol)
	}

	if protocol == "udp" {
		addr, err := net.ResolveUDPAddr("udp", address)
		if err != nil {
			return err
		}
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			return err
		}
		defer conn.Close()

		buf := make([]byte, 64*1024)
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}
		if verbose {
			fmt.Fprintf(stderr, "Connection from %s\n", remoteAddr)
		}

		// Write initial data to stdout
		if _, err := stdout.Write(buf[:n]); err != nil {
			return err
		}

		// To send data back, we need a separate goroutine reading stdin
		go func() {
			io.Copy(&udpWriter{conn, remoteAddr}, stdin)
			conn.Close()
		}()

		// Continue reading from socket
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				// check if closed
				if strings.Contains(err.Error(), "use of closed network connection") {
					return nil
				}
				return err
			}
			if _, err := stdout.Write(buf[:n]); err != nil {
				return err
			}
		}

	} else {
		ln, err := net.Listen(protocol, address)
		if err != nil {
			return err
		}
		defer ln.Close()

		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		defer conn.Close()

		if verbose {
			fmt.Fprintf(stderr, "Connection from %s\n", conn.RemoteAddr())
		}

		return pipeStream(conn, stdin, stdout)
	}
}

func pipeStream(conn net.Conn, stdin io.Reader, stdout io.Writer) error {
	// Write to connection in background
	go func() {
		_, _ = io.Copy(conn, stdin)
		// Close write side of connection if possible (TCP close write)
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	}()

	// Read from connection (blocks until remote closes or error)
	_, err := io.Copy(stdout, conn)
	return err
}

type udpWriter struct {
	conn *net.UDPConn
	addr *net.UDPAddr
}

func (w *udpWriter) Write(p []byte) (n int, err error) {
	return w.conn.WriteToUDP(p, w.addr)
}
