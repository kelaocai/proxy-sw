package macos

import (
	"fmt"
	"strings"
	"testing"
)

type fakeRunner struct {
	outputs map[string]string
	errs    map[string]error
	calls   []string
}

func (f *fakeRunner) Run(name string, args ...string) (string, error) {
	key := name + " " + strings.Join(args, " ")
	f.calls = append(f.calls, key)
	if err, ok := f.errs[key]; ok {
		return f.outputs[key], err
	}
	return f.outputs[key], nil
}

func TestListServices(t *testing.T) {
	runner := &fakeRunner{
		outputs: map[string]string{
			"networksetup -listallnetworkservices": "An asterisk (*) denotes that a network service is disabled.\nWi-Fi\n*Thunderbolt Bridge\nUSB 10/100/1000 LAN\n",
		},
	}
	manager := NewManager(runner)
	services, err := manager.ListServices()
	if err != nil {
		t.Fatalf("ListServices() error = %v", err)
	}
	got := strings.Join(services, ",")
	want := "Wi-Fi,USB 10/100/1000 LAN"
	if got != want {
		t.Fatalf("ListServices() = %q, want %q", got, want)
	}
}

func TestListServicesUnavailable(t *testing.T) {
	runner := &fakeRunner{
		outputs: map[string]string{
			"networksetup -listallnetworkservices": "AuthorizationCreate() failed: -60008",
		},
	}
	manager := NewManager(runner)
	_, err := manager.ListServices()
	if err == nil {
		t.Fatal("ListServices() error = nil, want authorization failure")
	}
}

func TestStatusUnavailable(t *testing.T) {
	runner := &fakeRunner{
		outputs: map[string]string{
			"networksetup -getwebproxy Wi-Fi":           "AuthorizationCreate() failed: -60008",
			"networksetup -getsecurewebproxy Wi-Fi":     "Enabled: Yes\nServer: 127.0.0.1\nPort: 7897\n",
			"networksetup -getsocksfirewallproxy Wi-Fi": "Enabled: No\nServer: 127.0.0.1\nPort: 7897\n",
			"networksetup -getproxybypassdomains Wi-Fi": "localhost\n*.home.arpa\n",
		},
		errs: map[string]error{
			"networksetup -getwebproxy Wi-Fi": fmt.Errorf("exit status 1"),
		},
	}
	manager := NewManager(runner)
	status, err := manager.Status("Wi-Fi")
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.Web.Available {
		t.Fatal("web proxy should be unavailable")
	}
	if !status.HTTPS.Enabled {
		t.Fatal("https proxy should be enabled")
	}
	if status.SOCKS.Enabled {
		t.Fatal("socks proxy should be disabled")
	}
	if got := strings.Join(status.BypassDomains, ","); got != "localhost,*.home.arpa" {
		t.Fatalf("unexpected bypass domains: %q", got)
	}
}

func TestEnableRunsAllCommands(t *testing.T) {
	runner := &fakeRunner{outputs: map[string]string{}}
	manager := NewManager(runner)
	if err := manager.Enable("Wi-Fi", "127.0.0.1", 7897, []string{"localhost", "*.home.arpa"}); err != nil {
		t.Fatalf("Enable() error = %v", err)
	}
	if len(runner.calls) != 7 {
		t.Fatalf("Enable() calls = %d, want 7", len(runner.calls))
	}
	if runner.calls[6] != "networksetup -setproxybypassdomains Wi-Fi localhost *.home.arpa" {
		t.Fatalf("unexpected bypass command: %q", runner.calls[6])
	}
}

func TestParseBypassDomains(t *testing.T) {
	got := parseBypassDomains("localhost\n*.home.arpa\nlocalhost\n")
	want := "localhost,*.home.arpa"
	if joined := strings.Join(got, ","); joined != want {
		t.Fatalf("parseBypassDomains() = %q, want %q", joined, want)
	}
}

func TestParseBypassDomainsEmpty(t *testing.T) {
	got := parseBypassDomains("There aren't any bypass domains set on Wi-Fi.\n")
	if len(got) != 0 {
		t.Fatalf("parseBypassDomains() = %+v, want empty", got)
	}
}
