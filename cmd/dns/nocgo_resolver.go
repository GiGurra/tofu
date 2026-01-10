//go:build !cgo || windows

package dns

import (
	"context"
	"net"
	"time"
)

// cgoEnabled indicates whether this build has CGO support
const cgoEnabled = false

// defaultTimeout for DNS lookups when no context is provided
const defaultTimeout = 2 * time.Second

// lookupHostCgo is a stub when CGO is not available.
// It falls back to Go's default resolver with a timeout.
func lookupHostCgo(hostname string) ([]net.IP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return net.DefaultResolver.LookupIP(ctx, "ip", hostname)
}

// MXRecordCgo represents an MX record
type MXRecordCgo struct {
	Pref uint16
	Host string
}

// lookupCNAMECgo falls back to Go's resolver with a timeout
func lookupCNAMECgo(hostname string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return net.DefaultResolver.LookupCNAME(ctx, hostname)
}

// lookupMXCgo falls back to Go's resolver with a timeout
func lookupMXCgo(hostname string) ([]*MXRecordCgo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	mxs, err := net.DefaultResolver.LookupMX(ctx, hostname)
	if err != nil {
		return nil, err
	}
	var result []*MXRecordCgo
	for _, mx := range mxs {
		result = append(result, &MXRecordCgo{
			Pref: mx.Pref,
			Host: mx.Host,
		})
	}
	return result, nil
}

// lookupTXTCgo falls back to Go's resolver with a timeout
func lookupTXTCgo(hostname string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return net.DefaultResolver.LookupTXT(ctx, hostname)
}

// lookupNSCgo falls back to Go's resolver with a timeout
func lookupNSCgo(hostname string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	nss, err := net.DefaultResolver.LookupNS(ctx, hostname)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, ns := range nss {
		result = append(result, ns.Host)
	}
	return result, nil
}
