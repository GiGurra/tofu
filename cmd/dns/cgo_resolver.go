//go:build cgo && !windows

package dns

/*
#cgo LDFLAGS: -lresolv
#include <stdlib.h>
#include <string.h>
#include <netdb.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <arpa/nameser.h>
#ifdef __APPLE__
#include <arpa/nameser_compat.h>
#endif
#include <resolv.h>

// Wrapper to lookup hostname and return IPs and canonical name
// Format: "canonical_name|ip1,ip2,ip3"
// Returns NULL on error, caller must free the result
char* cgo_getaddrinfo(const char* hostname) {
    struct addrinfo hints, *res, *p;
    char ipstr[INET6_ADDRSTRLEN];
    char* result = NULL;
    size_t result_len = 0;
    size_t result_cap = 512;

    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_STREAM;
    hints.ai_flags = AI_CANONNAME;

    int status = getaddrinfo(hostname, NULL, &hints, &res);
    if (status != 0) {
        return NULL;
    }

    result = (char*)malloc(result_cap);
    if (!result) {
        freeaddrinfo(res);
        return NULL;
    }
    result[0] = '\0';

    // First, add canonical name if available
    if (res->ai_canonname != NULL) {
        size_t cname_len = strlen(res->ai_canonname);
        if (cname_len + 2 > result_cap) {
            result_cap = (cname_len + 2) * 2;
            char* new_result = (char*)realloc(result, result_cap);
            if (!new_result) {
                free(result);
                freeaddrinfo(res);
                return NULL;
            }
            result = new_result;
        }
        strcpy(result, res->ai_canonname);
        result_len = cname_len;
    }
    // Add separator
    result[result_len++] = '|';
    result[result_len] = '\0';

    for (p = res; p != NULL; p = p->ai_next) {
        void* addr;
        if (p->ai_family == AF_INET) {
            struct sockaddr_in* ipv4 = (struct sockaddr_in*)p->ai_addr;
            addr = &(ipv4->sin_addr);
        } else if (p->ai_family == AF_INET6) {
            struct sockaddr_in6* ipv6 = (struct sockaddr_in6*)p->ai_addr;
            addr = &(ipv6->sin6_addr);
        } else {
            continue;
        }

        inet_ntop(p->ai_family, addr, ipstr, sizeof(ipstr));

        size_t ip_len = strlen(ipstr);
        size_t needed = result_len + ip_len + 2;

        if (needed > result_cap) {
            result_cap = needed * 2;
            char* new_result = (char*)realloc(result, result_cap);
            if (!new_result) {
                free(result);
                freeaddrinfo(res);
                return NULL;
            }
            result = new_result;
        }

        // Add comma if not first IP
        if (result[result_len - 1] != '|') {
            result[result_len++] = ',';
        }
        strcpy(result + result_len, ipstr);
        result_len += ip_len;
    }

    freeaddrinfo(res);
    return result;
}

// DNS record type constants
#define DNS_TYPE_MX 15
#define DNS_TYPE_TXT 16
#define DNS_TYPE_NS 2

// Helper to expand compressed DNS name
static int expand_name(const unsigned char* msg, const unsigned char* msgend,
                       const unsigned char* ptr, char* buf, size_t buflen) {
    int len = dn_expand(msg, msgend, ptr, buf, buflen);
    return len;
}

// Query DNS for specific record type using res_query
// Returns results as newline-separated strings, NULL on error
// For MX: "preference host" per line
// For TXT/NS: one record per line
char* cgo_res_query(const char* hostname, int type) {
    unsigned char answer[4096];
    char buf[1024];
    char* result = NULL;
    size_t result_len = 0;
    size_t result_cap = 1024;

    // Initialize resolver
    res_init();

    int len = res_query(hostname, C_IN, type, answer, sizeof(answer));
    if (len < 0) {
        return NULL;
    }

    // Parse the DNS response
    HEADER* header = (HEADER*)answer;
    int qdcount = ntohs(header->qdcount);
    int ancount = ntohs(header->ancount);

    if (ancount == 0) {
        return NULL;
    }

    result = (char*)malloc(result_cap);
    if (!result) return NULL;
    result[0] = '\0';

    const unsigned char* ptr = answer + sizeof(HEADER);
    const unsigned char* msgend = answer + len;

    // Skip question section
    for (int i = 0; i < qdcount; i++) {
        int n = dn_skipname(ptr, msgend);
        if (n < 0) {
            free(result);
            return NULL;
        }
        ptr += n + 4; // skip name + type + class
    }

    // Parse answer section
    for (int i = 0; i < ancount && ptr < msgend; i++) {
        // Skip name
        int n = dn_skipname(ptr, msgend);
        if (n < 0) break;
        ptr += n;

        if (ptr + 10 > msgend) break;

        // Read type, class, ttl, rdlength
        int rtype = (ptr[0] << 8) | ptr[1];
        // int rclass = (ptr[2] << 8) | ptr[3];
        // int ttl = (ptr[4] << 24) | (ptr[5] << 16) | (ptr[6] << 8) | ptr[7];
        int rdlength = (ptr[8] << 8) | ptr[9];
        ptr += 10;

        if (ptr + rdlength > msgend) break;

        if (rtype != type) {
            ptr += rdlength;
            continue;
        }

        char line[1100] = "";  // Large enough for buf (1024) + preference number

        if (type == DNS_TYPE_MX) {
            if (rdlength < 3) {
                ptr += rdlength;
                continue;
            }
            int pref = (ptr[0] << 8) | ptr[1];
            if (expand_name(answer, msgend, ptr + 2, buf, sizeof(buf)) > 0) {
                snprintf(line, sizeof(line), "%d %s", pref, buf);
            }
        } else if (type == DNS_TYPE_TXT) {
            // TXT records: one or more length-prefixed strings
            const unsigned char* rdata_end = ptr + rdlength;
            const unsigned char* rdata = ptr;
            size_t line_pos = 0;
            while (rdata < rdata_end && line_pos < sizeof(line) - 1) {
                int txt_len = *rdata++;
                if (rdata + txt_len > rdata_end) break;
                if (line_pos + txt_len >= sizeof(line)) break;
                memcpy(line + line_pos, rdata, txt_len);
                line_pos += txt_len;
                rdata += txt_len;
            }
            line[line_pos] = '\0';
        } else if (type == DNS_TYPE_NS) {
            if (expand_name(answer, msgend, ptr, buf, sizeof(buf)) > 0) {
                strncpy(line, buf, sizeof(line) - 1);
                line[sizeof(line) - 1] = '\0';
            }
        }

        ptr += rdlength;

        if (line[0] == '\0') continue;

        size_t line_len = strlen(line);
        size_t needed = result_len + line_len + 2;
        if (needed > result_cap) {
            result_cap = needed * 2;
            char* new_result = (char*)realloc(result, result_cap);
            if (!new_result) {
                free(result);
                return NULL;
            }
            result = new_result;
        }

        if (result_len > 0) {
            result[result_len++] = '\n';
        }
        strcpy(result + result_len, line);
        result_len += line_len;
    }

    if (result_len == 0) {
        free(result);
        return NULL;
    }

    return result;
}
*/
import "C"

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"unsafe"
)

// cgoEnabled indicates whether this build has CGO support
const cgoEnabled = true

// AddrInfoResult holds the result of a getaddrinfo call
type AddrInfoResult struct {
	CanonicalName string
	IPs           []net.IP
}

// lookupHostCgo uses the system's getaddrinfo via CGO to resolve a hostname.
// This ensures the lookup goes through libc, which tools like mirrord can intercept.
func lookupHostCgo(hostname string) ([]net.IP, error) {
	result, err := lookupAddrInfoCgo(hostname)
	if err != nil {
		return nil, err
	}
	return result.IPs, nil
}

// lookupAddrInfoCgo returns both IPs and canonical name
func lookupAddrInfoCgo(hostname string) (*AddrInfoResult, error) {
	cHostname := C.CString(hostname)
	defer C.free(unsafe.Pointer(cHostname))

	result := C.cgo_getaddrinfo(cHostname)
	if result == nil {
		return nil, errors.New("getaddrinfo failed: no such host")
	}
	defer C.free(unsafe.Pointer(result))

	resultStr := C.GoString(result)
	parts := strings.SplitN(resultStr, "|", 2)

	addrResult := &AddrInfoResult{}

	if len(parts) >= 1 && parts[0] != "" {
		addrResult.CanonicalName = parts[0]
	}

	if len(parts) >= 2 && parts[1] != "" {
		ipParts := strings.Split(parts[1], ",")
		seen := make(map[string]bool)

		for _, ipStr := range ipParts {
			if seen[ipStr] {
				continue
			}
			seen[ipStr] = true

			ip := net.ParseIP(ipStr)
			if ip != nil {
				addrResult.IPs = append(addrResult.IPs, ip)
			}
		}
	}

	if len(addrResult.IPs) == 0 {
		return nil, errors.New("no valid IP addresses found")
	}

	return addrResult, nil
}

// lookupCNAMECgo returns the canonical name for a hostname
func lookupCNAMECgo(hostname string) (string, error) {
	result, err := lookupAddrInfoCgo(hostname)
	if err != nil {
		return "", err
	}
	if result.CanonicalName == "" {
		return "", errors.New("no canonical name found")
	}
	return result.CanonicalName, nil
}

// MXRecordCgo represents an MX record
type MXRecordCgo struct {
	Pref uint16
	Host string
}

// lookupMXCgo queries MX records using res_query
func lookupMXCgo(hostname string) ([]*MXRecordCgo, error) {
	cHostname := C.CString(hostname)
	defer C.free(unsafe.Pointer(cHostname))

	result := C.cgo_res_query(cHostname, C.DNS_TYPE_MX)
	if result == nil {
		return nil, errors.New("no MX records found")
	}
	defer C.free(unsafe.Pointer(result))

	resultStr := C.GoString(result)
	lines := strings.Split(resultStr, "\n")

	var records []*MXRecordCgo
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		pref, err := strconv.ParseUint(parts[0], 10, 16)
		if err != nil {
			continue
		}
		records = append(records, &MXRecordCgo{
			Pref: uint16(pref),
			Host: parts[1],
		})
	}

	if len(records) == 0 {
		return nil, errors.New("no MX records found")
	}
	return records, nil
}

// lookupTXTCgo queries TXT records using res_query
func lookupTXTCgo(hostname string) ([]string, error) {
	cHostname := C.CString(hostname)
	defer C.free(unsafe.Pointer(cHostname))

	result := C.cgo_res_query(cHostname, C.DNS_TYPE_TXT)
	if result == nil {
		return nil, errors.New("no TXT records found")
	}
	defer C.free(unsafe.Pointer(result))

	resultStr := C.GoString(result)
	lines := strings.Split(resultStr, "\n")

	var records []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			records = append(records, line)
		}
	}

	if len(records) == 0 {
		return nil, errors.New("no TXT records found")
	}
	return records, nil
}

// lookupNSCgo queries NS records using res_query
func lookupNSCgo(hostname string) ([]string, error) {
	cHostname := C.CString(hostname)
	defer C.free(unsafe.Pointer(cHostname))

	result := C.cgo_res_query(cHostname, C.DNS_TYPE_NS)
	if result == nil {
		return nil, errors.New("no NS records found")
	}
	defer C.free(unsafe.Pointer(result))

	resultStr := C.GoString(result)
	lines := strings.Split(resultStr, "\n")

	var records []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			records = append(records, line)
		}
	}

	if len(records) == 0 {
		return nil, errors.New("no NS records found")
	}
	return records, nil
}
