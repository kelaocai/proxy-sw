package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kelaocai/proxy-sw/internal/network"
)

func TestExitCode(t *testing.T) {
	err := exitError{code: 2, err: testErr("bad input")}
	if got := ExitCode(err); got != 2 {
		t.Fatalf("ExitCode() = %d, want 2", got)
	}
}

func TestHelpShowsTopLevelListAndUse(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	app := New("dev", &stdout, &stderr)
	_ = app.Run([]string{"help"})
	got := stdout.String()
	if !strings.Contains(got, "  list") {
		t.Fatalf("help missing list command: %s", got)
	}
	if !strings.Contains(got, "  use NAME") {
		t.Fatalf("help missing use command: %s", got)
	}
	if !strings.Contains(got, "  doctor") {
		t.Fatalf("help missing doctor command: %s", got)
	}
	if !strings.Contains(got, "  detect") {
		t.Fatalf("help missing detect command: %s", got)
	}
	if !strings.Contains(got, "  system on|off|status") {
		t.Fatalf("help missing system command: %s", got)
	}
	if !strings.Contains(got, "  set --http-host HOST --http-port PORT [--socks5-host HOST --socks5-port PORT] [--no-proxy-add VALUES] [--no-proxy-clear-custom]") {
		t.Fatalf("help missing updated set usage: %s", got)
	}
	if strings.Contains(got, "service list") || strings.Contains(got, "service use") {
		t.Fatalf("help still contains legacy service commands: %s", got)
	}
}

func TestPreserveCustomNoProxy(t *testing.T) {
	networks := []network.LocalNetwork{{NetworkCIDR: "192.168.2.0/24"}}
	got := preserveCustomNoProxy(
		"localhost,127.0.0.1,192.168.2.0/24,custom.internal,*.home.arpa",
		networks,
		[]string{"preconfigured.local", "*.home.arpa"},
	)
	want := []string{"preconfigured.local", "*.home.arpa", "custom.internal"}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d (%+v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

type testErr string

func (e testErr) Error() string { return string(e) }
