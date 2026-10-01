package probe

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/miekg/dns"
	"github.com/octoplorer/octopulse/internal/domain"
)

func (x *execution) dns(ctx context.Context, c domain.DNSConfig) (Result, error) {
	result := Result{}
	recordType := dns.StringToType[strings.ToUpper(c.RecordType)]
	if recordType == 0 {
		return result, errors.New("unsupported DNS record type")
	}
	server := c.Server
	if server == "" {
		config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
		if err != nil || len(config.Servers) == 0 {
			return result, errors.New("no system DNS server; configure an explicit resolver")
		}
		server = net.JoinHostPort(config.Servers[0], config.Port)
	}
	name := c.Name
	if recordType == dns.TypePTR && net.ParseIP(name) != nil {
		var err error
		name, err = dns.ReverseAddr(name)
		if err != nil {
			return result, errors.New("invalid reverse DNS target")
		}
	}
	message := new(dns.Msg)
	message.SetQuestion(dns.Fqdn(name), recordType)
	message.SetEdns0(1232, false)
	client := &dns.Client{Net: c.Protocol}
	response, _, err := client.ExchangeContext(ctx, message, server)
	if err != nil {
		return result, errors.New("DNS exchange failed")
	}
	if response.Truncated && c.Protocol == "udp" {
		client.Net = "tcp"
		response, _, err = client.ExchangeContext(ctx, message, server)
		if err != nil {
			return result, errors.New("truncated DNS TCP fallback failed")
		}
	}
	if response.Truncated {
		return result, errors.New("DNS response is truncated")
	}
	rcode := dns.RcodeToString[response.Rcode]
	result.Diagnostics.DNSRCode = rcode
	values := []string{}
	for _, answer := range response.Answer {
		if answer.Header().Rrtype == recordType {
			values = append(values, dnsValue(answer))
		}
	}
	sort.Strings(values)
	result.Diagnostics.DNSValues = values
	if rcode != strings.ToUpper(c.ExpectedRCode) {
		return result, errors.New("DNS response code assertion failed")
	}
	expected := append([]string{}, c.ExpectedValues...)
	sort.Strings(expected)
	if c.MatchMode == "exact" && len(values) != len(expected) {
		return result, errors.New("DNS record set assertion failed")
	}
	for i, value := range expected {
		matched := false
		for _, actual := range values {
			if normalizedDNSValue(recordType, value) == normalizedDNSValue(recordType, actual) {
				matched = true
				break
			}
		}
		if !matched {
			return result, assertionError("DNS record", i)
		}
	}
	result.Success = true
	return result, nil
}

func normalizedDNSValue(recordType uint16, value string) string {
	switch recordType {
	case dns.TypeCNAME, dns.TypeNS, dns.TypePTR:
		return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(value), "."))
	case dns.TypeA, dns.TypeAAAA:
		if ip := net.ParseIP(strings.TrimSpace(value)); ip != nil {
			return ip.String()
		}
	}
	return strings.TrimSpace(value)
}
func dnsValue(answer dns.RR) string {
	switch r := answer.(type) {
	case *dns.A:
		return r.A.String()
	case *dns.AAAA:
		return r.AAAA.String()
	case *dns.CNAME:
		return r.Target
	case *dns.MX:
		return fmt.Sprintf("%d %s", r.Preference, r.Mx)
	case *dns.TXT:
		return strings.Join(r.Txt, "")
	case *dns.NS:
		return r.Ns
	case *dns.SRV:
		return fmt.Sprintf("%d %d %d %s", r.Priority, r.Weight, r.Port, r.Target)
	case *dns.PTR:
		return r.Ptr
	case *dns.SOA:
		return fmt.Sprintf("%s %s %d %d %d %d %d", r.Ns, r.Mbox, r.Serial, r.Refresh, r.Retry, r.Expire, r.Minttl)
	case *dns.CAA:
		return fmt.Sprintf("%d %s %s", r.Flag, r.Tag, r.Value)
	}
	return ""
}
