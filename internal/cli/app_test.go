package cli

import (
	"bytes"
	"strings"
	"testing"
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
	if strings.Contains(got, "service list") || strings.Contains(got, "service use") {
		t.Fatalf("help still contains legacy service commands: %s", got)
	}
}

type testErr string

func (e testErr) Error() string { return string(e) }
