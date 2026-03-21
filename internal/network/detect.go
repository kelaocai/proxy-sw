package network

import (
	"fmt"
	"net"
	"sort"
	"strings"
)

type LocalNetwork struct {
	Interface   string
	IPAddress   string
	NetworkCIDR string
}

type InterfaceProvider interface {
	Interfaces() ([]net.Interface, error)
	Addrs(net.Interface) ([]net.Addr, error)
}

type OSProvider struct{}

func (OSProvider) Interfaces() ([]net.Interface, error) {
	return net.Interfaces()
}

func (OSProvider) Addrs(iface net.Interface) ([]net.Addr, error) {
	return iface.Addrs()
}

func DetectLocalNetworks() ([]LocalNetwork, error) {
	return DetectLocalNetworksWith(OSProvider{})
}

func DetectLocalNetworksWith(provider InterfaceProvider) ([]LocalNetwork, error) {
	ifaces, err := provider.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("list interfaces: %w", err)
	}
	seen := map[string]bool{}
	networks := make([]LocalNetwork, 0)
	for _, iface := range ifaces {
		addrs, err := provider.Addrs(iface)
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet == nil {
				continue
			}
			ip4 := ipnet.IP.To4()
			if ip4 == nil || !isPrivateIPv4(ip4) {
				continue
			}
			ones, bits := ipnet.Mask.Size()
			if bits == 32 && ones >= 31 {
				continue
			}
			cidr := (&net.IPNet{IP: ip4.Mask(ipnet.Mask), Mask: ipnet.Mask}).String()
			key := iface.Name + "|" + cidr + "|" + ip4.String()
			if seen[key] {
				continue
			}
			seen[key] = true
			networks = append(networks, LocalNetwork{
				Interface:   iface.Name,
				IPAddress:   ip4.String(),
				NetworkCIDR: cidr,
			})
		}
	}
	sort.Slice(networks, func(i, j int) bool {
		if networks[i].Interface == networks[j].Interface {
			if networks[i].NetworkCIDR == networks[j].NetworkCIDR {
				return networks[i].IPAddress < networks[j].IPAddress
			}
			return networks[i].NetworkCIDR < networks[j].NetworkCIDR
		}
		return networks[i].Interface < networks[j].Interface
	})
	return networks, nil
}

func BaseNoProxyList() []string {
	return []string{
		"localhost",
		"127.0.0.1",
		"::1",
		"*.local",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}
}

func GenerateNoProxyList(networks []LocalNetwork, custom []string) []string {
	base := BaseNoProxyList()
	seen := map[string]bool{}
	out := make([]string, 0, len(base)+len(networks)+len(custom))
	appendUnique := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			return
		}
		seen[value] = true
		out = append(out, value)
	}
	for _, value := range base {
		appendUnique(value)
	}
	for _, n := range networks {
		appendUnique(n.NetworkCIDR)
	}
	for _, value := range custom {
		appendUnique(value)
	}
	return out
}

func ParseNoProxyCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || seen[part] {
			continue
		}
		seen[part] = true
		out = append(out, part)
	}
	return out
}

func UserCustomNoProxy(existing []string, networks []LocalNetwork) []string {
	auto := GenerateNoProxyList(networks, nil)
	autoSet := make(map[string]struct{}, len(auto))
	for _, value := range auto {
		autoSet[value] = struct{}{}
	}
	out := make([]string, 0, len(existing))
	seen := map[string]bool{}
	for _, value := range existing {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		if _, ok := autoSet[value]; ok {
			continue
		}
		out = append(out, value)
	}
	return out
}

func isPrivateIPv4(ip net.IP) bool {
	switch {
	case ip[0] == 10:
		return true
	case ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31:
		return true
	case ip[0] == 192 && ip[1] == 168:
		return true
	default:
		return false
	}
}
