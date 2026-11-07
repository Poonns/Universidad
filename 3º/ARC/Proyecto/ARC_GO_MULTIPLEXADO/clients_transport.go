package protocol

import "net/netip"

// Empty configuration is backward-compatible canonical-only mode. An explicit
// list REPLACES that default; the ingress is not implicitly added to the list.
func clientServerSources(ingress netip.AddrPort, configured []netip.AddrPort) ([3]netip.AddrPort, int, error) {
	var out [3]netip.AddrPort
	if len(configured) == 0 {
		out[0] = ingress
		return out, 1, nil
	}
	if len(configured) > len(out) {
		return out, 0, ErrClientConfig
	}
	for i, e := range configured {
		if !e.IsValid() {
			return out, 0, ErrClientConfig
		}
		e = normalizeEndpoint(e)
		if e.Port() == 0 || e.Addr().IsUnspecified() || e.Addr().IsMulticast() {
			return out, 0, ErrClientConfig
		}
		for j := 0; j < i; j++ {
			if out[j] == e {
				return out, 0, ErrClientConfig
			}
		}
		out[i] = e
	}
	return out, len(configured), nil
}

func (g *ClientGenerator) acceptsServerSource(source netip.AddrPort) bool {
	if !source.IsValid() {
		return false
	}
	source = normalizeEndpoint(source)
	for i := 0; i < g.serverSourceCount; i++ {
		if source == g.serverSources[i] {
			return true
		}
	}
	return false
}
