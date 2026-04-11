package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	ManagedStart = "# === proxy-sw managed start ==="
	ManagedEnd   = "# === proxy-sw managed end ==="
)

type Type string

const (
	Zsh  Type = "zsh"
	Bash Type = "bash"
	Fish Type = "fish"
)

type Env struct {
	HTTPHost  string
	HTTPPort  int
	SOCKSHost string
	SOCKSPort int
	NoProxy   []string
}

type Status struct {
	ShellType   Type
	Path        string
	BlockExists bool
	Values      map[string]string
}

type Manager struct{}

func NewManager() Manager { return Manager{} }

func (Manager) Detect(shellEnv string) (Type, error) {
	base := filepath.Base(shellEnv)
	switch base {
	case "zsh":
		return Zsh, nil
	case "bash":
		return Bash, nil
	case "fish":
		return Fish, nil
	default:
		if shellEnv == "" {
			return "", fmt.Errorf("unable to detect shell from environment")
		}
		return "", fmt.Errorf("unsupported shell %q", base)
	}
}

func (Manager) ConfigPath(shellType Type) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	switch shellType {
	case Zsh:
		return filepath.Join(home, ".zshrc"), nil
	case Bash:
		return filepath.Join(home, ".bashrc"), nil
	case Fish:
		return filepath.Join(home, ".config", "fish", "config.fish"), nil
	default:
		return "", fmt.Errorf("unsupported shell %q", shellType)
	}
}

func (Manager) Enable(path string, shellType Type, env Env) error {
	content, err := readOrEmpty(path)
	if err != nil {
		return err
	}
	block := buildBlock(shellType, env)
	updated := replaceManagedBlock(content, block)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create shell config dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write shell config: %w", err)
	}
	return nil
}

func (Manager) Disable(path string) error {
	content, err := readOrEmpty(path)
	if err != nil {
		return err
	}
	updated := removeManagedBlock(content)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create shell config dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write shell config: %w", err)
	}
	return nil
}

func (Manager) Status(path string, shellType Type) (Status, error) {
	content, err := readOrEmpty(path)
	if err != nil {
		return Status{}, err
	}
	block, exists := extractManagedBlock(content)
	status := Status{
		ShellType:   shellType,
		Path:        path,
		BlockExists: exists,
		Values:      map[string]string{},
	}
	if exists {
		status.Values = parseBlockValues(block, shellType)
	}
	return status, nil
}

func buildBlock(shellType Type, env Env) string {
	if env.HTTPHost == "" || env.HTTPPort <= 0 {
		return ""
	}
	socksHost := env.SOCKSHost
	socksPort := env.SOCKSPort
	if socksHost == "" || socksPort <= 0 {
		socksHost = env.HTTPHost
		socksPort = env.HTTPPort
	}
	httpURL := fmt.Sprintf("http://%s:%d", env.HTTPHost, env.HTTPPort)
	socksURL := fmt.Sprintf("socks5://%s:%d", socksHost, socksPort)
	noProxy := strings.Join(env.NoProxy, ",")
	lines := []string{ManagedStart}
	switch shellType {
	case Fish:
		lines = append(lines,
			fmt.Sprintf(`set -gx http_proxy "%s"`, httpURL),
			fmt.Sprintf(`set -gx https_proxy "%s"`, httpURL),
			fmt.Sprintf(`set -gx all_proxy "%s"`, socksURL),
			fmt.Sprintf(`set -gx HTTP_PROXY "%s"`, httpURL),
			fmt.Sprintf(`set -gx HTTPS_PROXY "%s"`, httpURL),
			fmt.Sprintf(`set -gx ALL_PROXY "%s"`, socksURL),
			fmt.Sprintf(`set -gx no_proxy "%s"`, noProxy),
			fmt.Sprintf(`set -gx NO_PROXY "%s"`, noProxy),
		)
	default:
		lines = append(lines,
			fmt.Sprintf(`export http_proxy="%s"`, httpURL),
			fmt.Sprintf(`export https_proxy="%s"`, httpURL),
			fmt.Sprintf(`export all_proxy="%s"`, socksURL),
			fmt.Sprintf(`export HTTP_PROXY="%s"`, httpURL),
			fmt.Sprintf(`export HTTPS_PROXY="%s"`, httpURL),
			fmt.Sprintf(`export ALL_PROXY="%s"`, socksURL),
			fmt.Sprintf(`export no_proxy="%s"`, noProxy),
			fmt.Sprintf(`export NO_PROXY="%s"`, noProxy),
		)
	}
	lines = append(lines, ManagedEnd)
	return strings.Join(lines, "\n")
}

func replaceManagedBlock(content, block string) string {
	if strings.TrimSpace(content) == "" {
		return block + "\n"
	}
	updated := removeManagedBlock(content)
	updated = strings.TrimRight(updated, "\n")
	if updated != "" {
		updated += "\n\n"
	}
	return updated + block + "\n"
}

func removeManagedBlock(content string) string {
	start := strings.Index(content, ManagedStart)
	end := strings.Index(content, ManagedEnd)
	if start == -1 || end == -1 || end < start {
		return content
	}
	end += len(ManagedEnd)
	if end < len(content) && content[end] == '\n' {
		end++
	}
	updated := content[:start] + content[end:]
	return strings.TrimLeft(updated, "\n")
}

func extractManagedBlock(content string) (string, bool) {
	start := strings.Index(content, ManagedStart)
	end := strings.Index(content, ManagedEnd)
	if start == -1 || end == -1 || end < start {
		return "", false
	}
	end += len(ManagedEnd)
	return content[start:end], true
}

func parseBlockValues(block string, shellType Type) map[string]string {
	values := map[string]string{}
	for _, raw := range strings.Split(block, "\n") {
		line := strings.TrimSpace(raw)
		switch shellType {
		case Fish:
			if !strings.HasPrefix(line, "set -gx ") {
				continue
			}
			rest := strings.TrimPrefix(line, "set -gx ")
			parts := strings.SplitN(rest, " ", 2)
			if len(parts) != 2 {
				continue
			}
			values[parts[0]] = strings.Trim(parts[1], `"`)
		default:
			if !strings.HasPrefix(line, "export ") {
				continue
			}
			rest := strings.TrimPrefix(line, "export ")
			key, value, ok := strings.Cut(rest, "=")
			if !ok {
				continue
			}
			values[key] = strings.Trim(value, `"`)
		}
	}
	return values
}

func readOrEmpty(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read shell config: %w", err)
	}
	return string(data), nil
}
