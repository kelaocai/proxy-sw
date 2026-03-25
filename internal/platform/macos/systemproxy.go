package macos

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Runner interface {
	Run(name string, args ...string) (string, error)
}

type RealRunner struct{}

func (RealRunner) Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return string(out), nil
}

type ProxyState struct {
	Available bool
	Enabled   bool
	Server    string
	Port      int
}

type Status struct {
	NetworkService string
	Web            ProxyState
	HTTPS          ProxyState
	SOCKS          ProxyState
	BypassDomains  []string
}

type Manager struct {
	runner Runner
}

func NewManager(r Runner) Manager {
	return Manager{runner: r}
}

func (m Manager) ListServices() ([]string, error) {
	out, err := m.runner.Run("networksetup", "-listallnetworkservices")
	if err != nil {
		return nil, err
	}
	if looksUnavailable(out) {
		return nil, errors.New(strings.TrimSpace(out))
	}
	lines := strings.Split(out, "\n")
	services := make([]string, 0, len(lines))
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "An asterisk") {
			continue
		}
		if strings.HasPrefix(line, "*") {
			continue
		}
		services = append(services, line)
	}
	if len(services) == 0 {
		return nil, errors.New("no enabled network services found")
	}
	return services, nil
}

func (m Manager) DetectService() (string, error) {
	services, err := m.ListServices()
	if err != nil {
		return "", err
	}
	return services[0], nil
}

func (m Manager) Enable(service, host string, port int, bypassDomains []string) error {
	commands := [][]string{
		{"-setwebproxy", service, host, strconv.Itoa(port)},
		{"-setsecurewebproxy", service, host, strconv.Itoa(port)},
		{"-setsocksfirewallproxy", service, host, strconv.Itoa(port)},
		{"-setwebproxystate", service, "on"},
		{"-setsecurewebproxystate", service, "on"},
		{"-setsocksfirewallproxystate", service, "on"},
	}
	if len(bypassDomains) == 0 {
		commands = append(commands, []string{"-setproxybypassdomains", service, "Empty"})
	} else {
		commands = append(commands, append([]string{"-setproxybypassdomains", service}, bypassDomains...))
	}
	return m.runBatch(commands)
}

func (m Manager) Disable(service string) error {
	commands := [][]string{
		{"-setwebproxystate", service, "off"},
		{"-setsecurewebproxystate", service, "off"},
		{"-setsocksfirewallproxystate", service, "off"},
	}
	return m.runBatch(commands)
}

func (m Manager) Status(service string) (Status, error) {
	web, err := m.readProxy(service, "-getwebproxy")
	if err != nil {
		return Status{}, err
	}
	https, err := m.readProxy(service, "-getsecurewebproxy")
	if err != nil {
		return Status{}, err
	}
	socks, err := m.readProxy(service, "-getsocksfirewallproxy")
	if err != nil {
		return Status{}, err
	}
	bypassDomains, err := m.readBypassDomains(service)
	if err != nil {
		return Status{}, err
	}
	return Status{
		NetworkService: service,
		Web:            web,
		HTTPS:          https,
		SOCKS:          socks,
		BypassDomains:  bypassDomains,
	}, nil
}

func (m Manager) runBatch(commands [][]string) error {
	for _, args := range commands {
		if _, err := m.runner.Run("networksetup", args...); err != nil {
			return err
		}
	}
	return nil
}

func (m Manager) readProxy(service, subcommand string) (ProxyState, error) {
	out, err := m.runner.Run("networksetup", subcommand, service)
	if err != nil {
		if looksUnavailable(out) {
			return ProxyState{Available: false}, nil
		}
		return ProxyState{}, err
	}
	return parseProxyState(out), nil
}

func (m Manager) readBypassDomains(service string) ([]string, error) {
	out, err := m.runner.Run("networksetup", "-getproxybypassdomains", service)
	if err != nil {
		if looksUnavailable(out) {
			return nil, nil
		}
		return nil, err
	}
	return parseBypassDomains(out), nil
}

func parseProxyState(out string) ProxyState {
	state := ProxyState{Available: true}
	for _, raw := range strings.Split(out, "\n") {
		line := strings.TrimSpace(raw)
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "Enabled":
			state.Enabled = value == "Yes"
		case "Server":
			state.Server = value
		case "Port":
			port, _ := strconv.Atoi(value)
			state.Port = port
		}
	}
	return state
}

func looksUnavailable(out string) bool {
	lower := strings.ToLower(out)
	return strings.Contains(lower, "authorization") ||
		strings.Contains(lower, "not authorized") ||
		strings.Contains(lower, "failed")
}

func parseBypassDomains(out string) []string {
	lines := strings.Split(out, "\n")
	domains := make([]string, 0, len(lines))
	seen := map[string]bool{}
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "aren't any bypass domains") {
			return nil
		}
		if seen[line] {
			continue
		}
		seen[line] = true
		domains = append(domains, line)
	}
	return domains
}
