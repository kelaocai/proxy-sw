package network

import (
	"net"
	"testing"
)

type fakeProvider struct {
	ifaces []net.Interface
	addrs  map[string][]net.Addr
}

func (f fakeProvider) Interfaces() ([]net.Interface, error) { return f.ifaces, nil }
func (f fakeProvider) Addrs(iface net.Interface) ([]net.Addr, error) {
	return f.addrs[iface.Name], nil
}

func TestDetectLocalNetworksWith(t *testing.T) {
	provider := fakeProvider{
		ifaces: []net.Interface{{Name: "en0"}, {Name: "utun0"}},
		addrs: map[string][]net.Addr{
			"en0": {
				&net.IPNet{IP: net.ParseIP("192.168.2.10"), Mask: net.CIDRMask(24, 32)},
				&net.IPNet{IP: net.ParseIP("8.8.8.8"), Mask: net.CIDRMask(24, 32)},
			},
			"utun0": {
				&net.IPNet{IP: net.ParseIP("10.0.1.5"), Mask: net.CIDRMask(24, 32)},
			},
		},
	}
	got, err := DetectLocalNetworksWith(provider)
	if err != nil {
		t.Fatalf("DetectLocalNetworksWith() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(networks) = %d, want 2", len(got))
	}
	if got[0].NetworkCIDR != "192.168.2.0/24" && got[1].NetworkCIDR != "192.168.2.0/24" {
		t.Fatalf("missing expected network: %+v", got)
	}
}

func TestGenerateNoProxyList(t *testing.T) {
	networks := []LocalNetwork{{NetworkCIDR: "192.168.2.0/24"}, {NetworkCIDR: "10.0.1.0/24"}}
	got := GenerateNoProxyList(networks, []string{"internal.example.com", "10.0.1.0/24"})
	if got[0] != "localhost" {
		t.Fatalf("unexpected first item: %+v", got)
	}
	foundCustom := false
	foundCIDR := false
	countCIDR := 0
	for _, value := range got {
		if value == "internal.example.com" {
			foundCustom = true
		}
		if value == "10.0.1.0/24" {
			foundCIDR = true
			countCIDR++
		}
	}
	if !foundCustom || !foundCIDR || countCIDR != 1 {
		t.Fatalf("unexpected no_proxy list: %+v", got)
	}
}

func TestParseNoProxyCSV(t *testing.T) {
	got := ParseNoProxyCSV(" localhost, 127.0.0.1,localhost, ,*.home.arpa ")
	want := []string{"localhost", "127.0.0.1", "*.home.arpa"}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d (%+v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestUserCustomNoProxy(t *testing.T) {
	networks := []LocalNetwork{{NetworkCIDR: "192.168.2.0/24"}}
	got := UserCustomNoProxy([]string{
		"localhost",
		"192.168.2.0/24",
		"*.home.arpa",
		"nas.local",
		"*.home.arpa",
	}, networks)
	want := []string{"*.home.arpa", "nas.local"}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d (%+v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
