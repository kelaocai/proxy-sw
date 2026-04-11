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

type Proxy struct {
	Host string
	Port int
}

type Config struct {
	HTTP           Proxy
	SOCKS5         Proxy
	NetworkService string
	ShellType      string
	NoProxyCustom  []string
}

func Default() Config {
	return Config{
		HTTP: Proxy{
			Host: DefaultHost,
			Port: DefaultPort,
		},
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
	var currentSection string
	for scanner.Scan() {
		lineNo++
		rawLine := scanner.Text()
		line := strings.TrimSpace(rawLine)
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

		if strings.HasPrefix(rawLine, "  ") {
			if currentSection == "" {
				return Config{}, fmt.Errorf("unexpected nested config on line %d", lineNo)
			}
			key, value, ok := strings.Cut(line, ":")
			if !ok {
				return Config{}, fmt.Errorf("invalid config line %d", lineNo)
			}
			key = strings.TrimSpace(key)
			value = strings.Trim(strings.TrimSpace(value), `"'`)
			if value == "" {
				return Config{}, fmt.Errorf("%s.%s cannot be empty", currentSection, key)
			}
			if err := setProxyValue(&cfg, currentSection, key, value); err != nil {
				return Config{}, err
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
			currentSection = ""
			switch key {
			case "http", "socks5":
				currentList = ""
				currentSection = key
				continue
			case "no_proxy_custom":
				cfg.NoProxyCustom = nil
				continue
			default:
				return Config{}, fmt.Errorf("unknown config key %q", key)
			}
		}
		currentList = ""
		currentSection = ""
		switch key {
		case "host":
			if value == "" {
				return Config{}, fmt.Errorf("host cannot be empty")
			}
			cfg.HTTP.Host = value
		case "port":
			port, err := strconv.Atoi(value)
			if err != nil || port <= 0 {
				return Config{}, fmt.Errorf("invalid port %q", value)
			}
			cfg.HTTP.Port = port
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
	if cfg.HTTP.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	if cfg.HTTP.Port <= 0 {
		return fmt.Errorf("port must be greater than zero")
	}
	if cfg.SOCKS5.Host != "" && cfg.SOCKS5.Port <= 0 {
		return fmt.Errorf("socks5 port must be greater than zero")
	}
	if cfg.SOCKS5.Port > 0 && cfg.SOCKS5.Host == "" {
		return fmt.Errorf("socks5 host cannot be empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	var content strings.Builder
	content.WriteString("# proxy-sw configuration\n")
	content.WriteString("http:\n")
	content.WriteString(fmt.Sprintf("  host: %s\n", cfg.HTTP.Host))
	content.WriteString(fmt.Sprintf("  port: %d\n", cfg.HTTP.Port))
	if cfg.SOCKS5.Host != "" && cfg.SOCKS5.Port > 0 {
		content.WriteString("socks5:\n")
		content.WriteString(fmt.Sprintf("  host: %s\n", cfg.SOCKS5.Host))
		content.WriteString(fmt.Sprintf("  port: %d\n", cfg.SOCKS5.Port))
	}
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

func setProxyValue(cfg *Config, section, key, value string) error {
	proxy := &cfg.HTTP
	switch section {
	case "http":
		proxy = &cfg.HTTP
	case "socks5":
		proxy = &cfg.SOCKS5
	default:
		return fmt.Errorf("unknown config section %q", section)
	}
	switch key {
	case "host":
		proxy.Host = value
	case "port":
		port, err := strconv.Atoi(value)
		if err != nil || port <= 0 {
			return fmt.Errorf("invalid port %q", value)
		}
		proxy.Port = port
	default:
		return fmt.Errorf("unknown config key %q.%s", section, key)
	}
	return nil
}
