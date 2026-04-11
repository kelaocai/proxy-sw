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
	env := Env{
		HTTPHost:   "127.0.0.1",
		HTTPPort:   7897,
		SOCKSHost:  "127.0.0.1",
		SOCKSPort:  6153,
		NoProxy:    []string{"localhost", "127.0.0.1"},
	}
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

func TestEnableUsesIndependentSocksProxy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".zshrc")
	manager := NewManager()
	env := Env{
		HTTPHost:  "127.0.0.1",
		HTTPPort:  6152,
		SOCKSHost: "127.0.0.1",
		SOCKSPort: 6153,
		NoProxy:   []string{"localhost", "127.0.0.1"},
	}
	if err := manager.Enable(path, Zsh, env); err != nil {
		t.Fatalf("Enable() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	content := string(data)
	if !strings.Contains(content, `export http_proxy="http://127.0.0.1:6152"`) {
		t.Fatalf("http proxy not written correctly: %s", content)
	}
	if !strings.Contains(content, `export all_proxy="socks5://127.0.0.1:6153"`) {
		t.Fatalf("socks proxy not written correctly: %s", content)
	}
}

func TestEnableFallsBackToHTTPForSocksProxy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".zshrc")
	manager := NewManager()
	env := Env{
		HTTPHost: "127.0.0.1",
		HTTPPort: 6152,
		NoProxy:  []string{"localhost", "127.0.0.1"},
	}
	if err := manager.Enable(path, Zsh, env); err != nil {
		t.Fatalf("Enable() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), `export all_proxy="socks5://127.0.0.1:6152"`) {
		t.Fatalf("expected socks fallback to http endpoint: %s", string(data))
	}
}

func TestBuildBlockRequiresHTTPProxy(t *testing.T) {
	if got := buildBlock(Zsh, Env{SOCKSHost: "127.0.0.1", SOCKSPort: 6153}); got != "" {
		t.Fatalf("buildBlock() = %q, want empty string without http proxy", got)
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

func TestStatusParsesUppercaseNoProxy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".zshrc")
	content := ManagedStart + "\n" +
		`export NO_PROXY="localhost,127.0.0.1,*.home.arpa"` + "\n" +
		ManagedEnd + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	status, err := NewManager().Status(path, Zsh)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if got := status.Values["NO_PROXY"]; got != "localhost,127.0.0.1,*.home.arpa" {
		t.Fatalf("NO_PROXY = %q", got)
	}
}
