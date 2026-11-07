package main

import (
	"fmt"
	"net/netip"
	"strings"

	m "arc.local/multiplexado/internal/measurement"
)

func serverTransport(ready m.Ready, sockets int) (string, string, error) {
	ingress := ready.IngressEndpoint
	if ingress == "" {
		ingress = ready.Address
	}
	parse := func(value string) (netip.AddrPort, error) {
		e, err := netip.ParseAddrPort(value)
		if err != nil {
			return e, fmt.Errorf("invalid READY endpoint %q", value)
		}
		e = netip.AddrPortFrom(e.Addr().Unmap(), e.Port())
		if e.Port() == 0 || e.Addr().IsUnspecified() || e.Addr().IsMulticast() {
			return e, fmt.Errorf("invalid READY source %q", value)
		}
		return e, nil
	}
	in, err := parse(ingress)
	if err != nil {
		return "", "", err
	}
	if ready.Address != ingress {
		return "", "", fmt.Errorf("READY Address must equal canonical ingress")
	}
	if len(ready.ServerEgressEndpoints) == 0 && sockets == 1 {
		return ingress, "", nil
	} // historical binaries
	if len(ready.ServerEgressEndpoints) != sockets {
		return "", "", fmt.Errorf("READY egress count mismatch")
	}
	seen := make([]netip.AddrPort, 0, sockets)
	for _, value := range ready.ServerEgressEndpoints {
		e, err := parse(value)
		if err != nil {
			return "", "", err
		}
		for _, previous := range seen {
			if previous == e {
				return "", "", fmt.Errorf("duplicate READY egress")
			}
		}
		if (sockets == 1 && e != in) || (sockets > 1 && e == in) {
			return "", "", fmt.Errorf("READY violates shared/independent descriptor mode")
		}
		seen = append(seen, e)
	}
	return ingress, strings.Join(ready.ServerEgressEndpoints, ","), nil
}
