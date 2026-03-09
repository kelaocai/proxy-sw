package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kelaocai/proxy-sw/internal/network"
	"github.com/kelaocai/proxy-sw/internal/platform/macos"
	"github.com/kelaocai/proxy-sw/internal/shell"
)

func TestRendererStatusPlain(t *testing.T) {
	var out bytes.Buffer
	renderer := Renderer{Color: false}
	status := macos.Status{
		NetworkService: "Wi-Fi",
		Web:            macos.ProxyState{Available: true, Enabled: true, Server: "127.0.0.1", Port: 7897},
		HTTPS:          macos.ProxyState{Available: true, Enabled: false},
		SOCKS:          macos.ProxyState{Available: false},
	}
	if err := renderer.SystemStatus(&out, status, "127.0.0.1", 7897); err != nil {
		t.Fatalf("SystemStatus() error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"System Proxy  Wi-Fi",
		"web    on",
		"https  off",
		"socks  unavailable",
		"127.0.0.1:7897",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q: %s", want, got)
		}
	}
}

func TestRendererShellStatusPlain(t *testing.T) {
	var out bytes.Buffer
	renderer := Renderer{Color: false}
	status := shell.Status{
		ShellType:   shell.Zsh,
		Path:        "/tmp/.zshrc",
		BlockExists: true,
		Values: map[string]string{
			"http_proxy": "http://127.0.0.1:7897",
			"no_proxy":   "localhost,127.0.0.1",
		},
	}
	if err := renderer.ShellStatus(&out, status); err != nil {
		t.Fatalf("ShellStatus() error = %v", err)
	}
	got := out.String()
	for _, want := range []string{"Shell Proxy  zsh", "file", "/tmp/.zshrc", "http_proxy", "no_proxy"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q: %s", want, got)
		}
	}
}

func TestRendererDoctorPlain(t *testing.T) {
	var out bytes.Buffer
	renderer := Renderer{Color: false}
	checks := []Check{
		{Name: "platform", Status: "on", Details: "macOS supported"},
		{Name: "port", Status: "warn", Details: "127.0.0.1:7897 unreachable"},
	}
	if err := renderer.Doctor(&out, checks); err != nil {
		t.Fatalf("Doctor() error = %v", err)
	}
	got := out.String()
	for _, want := range []string{"Proxy Doctor", "platform", "macOS supported", "127.0.0.1:7897 unreachable"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q: %s", want, got)
		}
	}
}

func TestRendererDetectPlain(t *testing.T) {
	var out bytes.Buffer
	renderer := Renderer{Color: false}
	networks := []network.LocalNetwork{{Interface: "en0", IPAddress: "192.168.2.10", NetworkCIDR: "192.168.2.0/24"}}
	if err := renderer.Detect(&out, networks, []string{"localhost", "192.168.2.0/24"}); err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	got := out.String()
	for _, want := range []string{"Local Networks", "en0", "192.168.2.0/24", "no_proxy"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q: %s", want, got)
		}
	}
}
