package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnableReplaceAndDisable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".zshrc")
	if err := os.WriteFile(path, []byte("export PATH=$PATH\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	manager := NewManager()
	env := Env{Host: "127.0.0.1", Port: 7897, NoProxy: []string{"localhost", "127.0.0.1"}}
	if err := manager.Enable(path, Zsh, env); err != nil {
		t.Fatalf("Enable() error = %v", err)
	}
	if err := manager.Enable(path, Zsh, env); err != nil {
		t.Fatalf("Enable() second error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	content := string(data)
	if strings.Count(content, ManagedStart) != 1 {
		t.Fatalf("managed block count = %d, want 1", strings.Count(content, ManagedStart))
	}
	if err := manager.Disable(path); err != nil {
		t.Fatalf("Disable() error = %v", err)
	}
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(data), ManagedStart) {
		t.Fatalf("managed block still present: %s", string(data))
	}
}

func TestStatusParsesValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".zshrc")
	content := ManagedStart + "\n" +
		`export http_proxy="http://127.0.0.1:7897"` + "\n" +
		`export no_proxy="localhost,127.0.0.1"` + "\n" +
		ManagedEnd + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	status, err := NewManager().Status(path, Zsh)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if !status.BlockExists || status.Values["http_proxy"] == "" || status.Values["no_proxy"] == "" {
		t.Fatalf("unexpected status: %+v", status)
	}
}
