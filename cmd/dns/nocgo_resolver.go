//go:build !cgo || windows

package dns

import (
	"net"
)

// cgoEnabled indicates whether this build has CGO support
const cgoEnabled = false

// lookupHostCgo is a stub when CGO is not available.
// It falls back to Go's default resolver.
func lookupHostCgo(hostname string) ([]net.IP, error) {
	return net.LookupIP(hostname)
}

// MXRecordCgo represents an MX record
type MXRecordCgo struct {
	Pref uint16
	Host string
}

// lookupCNAMECgo falls back to Go's resolver
func lookupCNAMECgo(hostname string) (string, error) {
	return net.LookupCNAME(hostname)
}

// lookupMXCgo falls back to Go's resolver
func lookupMXCgo(hostname string) ([]*MXRecordCgo, error) {
	mxs, err := net.LookupMX(hostname)
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

// lookupTXTCgo falls back to Go's resolver
func lookupTXTCgo(hostname string) ([]string, error) {
	return net.LookupTXT(hostname)
}

// lookupNSCgo falls back to Go's resolver
func lookupNSCgo(hostname string) ([]string, error) {
	nss, err := net.LookupNS(hostname)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, ns := range nss {
		result = append(result, ns.Host)
	}
	return result, nil
}
