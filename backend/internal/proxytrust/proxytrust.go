// Package proxytrust decides whether a request's direct peer is a reverse
// proxy whose forwarding headers may be trusted.
package proxytrust

import (
	"fmt"
	"net"
	"strings"
)

var defaultEntries = []string{"127.0.0.1/32", "::1/128"}

// Trust holds the CIDRs of reverse proxies allowed to set X-Forwarded-For,
// CF-Connecting-IP and X-Forwarded-Proto.
type Trust struct {
	networks []*net.IPNet
	entries  []string
}

// New parses proxy CIDRs. An empty list keeps the safe default (loopback only).
func New(entries []string) (*Trust, error) {
	if len(entries) == 0 {
		entries = defaultEntries
	}

	trust := &Trust{}
	seen := make(map[string]bool)
	for _, raw := range entries {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if !strings.Contains(raw, "/") {
			ip := net.ParseIP(raw)
			if ip == nil {
				return nil, fmt.Errorf("invalid trusted proxy address %q", raw)
			}
			if ip.To4() != nil {
				raw += "/32"
			} else {
				raw += "/128"
			}
		}
		_, network, err := net.ParseCIDR(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy CIDR %q: %w", raw, err)
		}
		if seen[raw] {
			continue
		}
		seen[raw] = true
		trust.networks = append(trust.networks, network)
		trust.entries = append(trust.entries, raw)
	}

	if len(trust.networks) == 0 {
		trust.networks = append(trust.networks, mustNetwork(defaultEntries[0]), mustNetwork(defaultEntries[1]))
		trust.entries = append(trust.entries, defaultEntries...)
	}
	return trust, nil
}

// Entries returns normalized CIDRs, ready for gin.Engine.SetTrustedProxies.
func (t *Trust) Entries() []string {
	return append([]string(nil), t.entries...)
}

// ContainsRemote reports whether the direct peer (host:port) is a trusted proxy.
func (t *Trust) ContainsRemote(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err != nil {
		host = strings.TrimSpace(remoteAddr)
	}
	return t.Contains(net.ParseIP(host))
}

// Contains reports whether an IP belongs to a trusted proxy network.
func (t *Trust) Contains(ip net.IP) bool {
	if ip == nil {
		return false
	}
	for _, network := range t.networks {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func mustNetwork(cidr string) *net.IPNet {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}
	return network
}
