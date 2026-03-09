package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	DefaultHost = "127.0.0.1"
	DefaultPort = 7897
)

type Config struct {
	Host           string
	Port           int
	NetworkService string
	ShellType      string
	NoProxyCustom  []string
}

func Default() Config {
	return Config{
		Host: DefaultHost,
		Port: DefaultPort,
	}
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".config", "proxy-sw", "config.yaml"), nil
}

func Load(path string) (Config, error) {
	cfg := Default()
	data, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer data.Close()

	scanner := bufio.NewScanner(data)
	lineNo := 0
	var currentList string
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "- ") {
			if currentList == "" {
				return Config{}, fmt.Errorf("unexpected list item on line %d", lineNo)
			}
			value := strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "- ")), `"'`)
			switch currentList {
			case "no_proxy_custom":
				cfg.NoProxyCustom = append(cfg.NoProxyCustom, value)
			default:
				return Config{}, fmt.Errorf("unknown list key %q", currentList)
			}
			continue
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return Config{}, fmt.Errorf("invalid config line %d", lineNo)
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if value == "" {
			currentList = key
			switch key {
			case "no_proxy_custom":
				cfg.NoProxyCustom = nil
				continue
			default:
				return Config{}, fmt.Errorf("unknown config key %q", key)
			}
		}
		currentList = ""
		switch key {
		case "host":
			if value == "" {
				return Config{}, fmt.Errorf("host cannot be empty")
			}
			cfg.Host = value
		case "port":
			port, err := strconv.Atoi(value)
			if err != nil || port <= 0 {
				return Config{}, fmt.Errorf("invalid port %q", value)
			}
			cfg.Port = port
		case "network_service":
			cfg.NetworkService = value
		case "shell_type":
			cfg.ShellType = value
		default:
			return Config{}, fmt.Errorf("unknown config key %q", key)
		}
	}
	if err := scanner.Err(); err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if cfg.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	if cfg.Port <= 0 {
		return fmt.Errorf("port must be greater than zero")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	var content strings.Builder
	content.WriteString("# proxy-sw configuration\n")
	content.WriteString(fmt.Sprintf("host: %s\n", cfg.Host))
	content.WriteString(fmt.Sprintf("port: %d\n", cfg.Port))
	if cfg.NetworkService != "" {
		content.WriteString(fmt.Sprintf("network_service: %s\n", cfg.NetworkService))
	}
	if cfg.ShellType != "" {
		content.WriteString(fmt.Sprintf("shell_type: %s\n", cfg.ShellType))
	}
	if len(cfg.NoProxyCustom) > 0 {
		content.WriteString("no_proxy_custom:\n")
		for _, entry := range cfg.NoProxyCustom {
			content.WriteString(fmt.Sprintf("  - %s\n", entry))
		}
	}
	if err := os.WriteFile(path, []byte(content.String()), 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
