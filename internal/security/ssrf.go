package security

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

var privateCIDRs []*net.IPNet

func init() {
	blocks := []string{
		"127.0.0.0/8",    // loopback
		"10.0.0.0/8",      // private
		"172.16.0.0/12",   // private
		"192.168.0.0/16",  // private
		"169.254.0.0/16",  // link-local
		"::1/128",         // loopback IPv6
		"fc00::/7",        // unique-local IPv6
		"fe80::/10",       // link-local IPv6
		"0.0.0.0/8",       // current network
		"100.64.0.0/10",   // carrier-grade NAT
		"198.18.0.0/15",   // benchmark testing
	}
	for _, cidr := range blocks {
		_, block, err := net.ParseCIDR(cidr)
		if err == nil {
			privateCIDRs = append(privateCIDRs, block)
		}
	}
}

func ValidateExternalURL(rawURL string, allowedDomains []string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("scheme %s not allowed", u.Scheme)
	}

	hostname := u.Hostname()
	if hostname == "" {
		return fmt.Errorf("empty hostname")
	}

	if len(allowedDomains) > 0 && !domainAllowed(hostname, allowedDomains) {
		return fmt.Errorf("domain %s not in whitelist", hostname)
	}

	ips, err := net.LookupHost(hostname)
	if err != nil {
		return fmt.Errorf("unresolvable host: %w", err)
	}

	for _, ip := range ips {
		parsed := net.ParseIP(ip)
		if parsed == nil {
			continue
		}
		if isPrivateIP(parsed) {
			return fmt.Errorf("SSRF blocked: %s resolves to private IP %s", hostname, ip)
		}
	}

	return nil
}

func domainAllowed(hostname string, allowed []string) bool {
	for _, a := range allowed {
		if strings.EqualFold(hostname, a) {
			return true
		}
		if strings.HasPrefix(a, ".") && strings.HasSuffix(hostname, a) {
			return true
		}
	}
	return false
}

func isPrivateIP(ip net.IP) bool {
	for _, block := range privateCIDRs {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}
