package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadDefaultsWhenMissing(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTP.Host != DefaultHost || cfg.HTTP.Port != DefaultPort {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadAndSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	want := Config{
		HTTP: Proxy{
			Host: "10.0.0.2",
			Port: 9090,
		},
		SOCKS5: Proxy{
			Host: "10.0.0.3",
			Port: 9091,
		},
		NetworkService: "USB 10/100/1000 LAN",
		ShellType:      "zsh",
		NoProxyCustom:  []string{"internal.example.com", "*.corp.local"},
	}
	if err := Save(path, want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() = %+v, want %+v", got, want)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("config file is empty")
	}
}

func TestLoadListSyntax(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `http:
  host: 127.0.0.1
  port: 7897
socks5:
  host: 127.0.0.1
  port: 6153
shell_type: zsh
no_proxy_custom:
  - a.internal
  - "*.corp.local"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(got.NoProxyCustom) != 2 || got.NoProxyCustom[0] != "a.internal" || got.NoProxyCustom[1] != "*.corp.local" {
		t.Fatalf("unexpected no_proxy_custom: %+v", got.NoProxyCustom)
	}
	if got.SOCKS5.Port != 6153 {
		t.Fatalf("unexpected socks5 config: %+v", got.SOCKS5)
	}
}

func TestLoadLegacyHostPortIntoHTTP(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `host: 127.0.0.1
port: 6152
shell_type: zsh
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.HTTP.Host != "127.0.0.1" || got.HTTP.Port != 6152 {
		t.Fatalf("unexpected http proxy: %+v", got.HTTP)
	}
	if got.SOCKS5.Host != "" || got.SOCKS5.Port != 0 {
		t.Fatalf("legacy config should not populate socks5: %+v", got.SOCKS5)
	}
}
