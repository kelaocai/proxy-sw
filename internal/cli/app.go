package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/kelaocai/proxy-sw/internal/config"
	"github.com/kelaocai/proxy-sw/internal/network"
	"github.com/kelaocai/proxy-sw/internal/output"
	"github.com/kelaocai/proxy-sw/internal/platform/macos"
	"github.com/kelaocai/proxy-sw/internal/shell"
)

type exitError struct {
	code int
	err  error
}

func (e exitError) Error() string { return e.err.Error() }
func (e exitError) Unwrap() error { return e.err }

func ExitCode(err error) int {
	var ee exitError
	if errors.As(err, &ee) {
		return ee.code
	}
	return 1
}

type App struct {
	version      string
	stdout       io.Writer
	stderr       io.Writer
	system       macos.Manager
	shellManager shell.Manager
}

func New(version string, stdout, stderr io.Writer) App {
	return App{
		version:      version,
		stdout:       stdout,
		stderr:       stderr,
		system:       macos.NewManager(macos.RealRunner{}),
		shellManager: shell.NewManager(),
	}
}

func (a App) Run(args []string) error {
	if runtime.GOOS != "darwin" {
		return exitError{code: 1, err: errors.New("proxy-sw currently supports macOS only")}
	}
	if len(args) == 0 {
		a.printHelp()
		return nil
	}
	switch args[0] {
	case "on":
		return a.runOn(args[1:])
	case "off":
		return a.runOff(args[1:])
	case "status":
		return a.runStatus(args[1:])
	case "detect":
		return a.runDetect()
	case "set":
		return a.runSet(args[1:])
	case "list":
		return a.runList()
	case "use":
		return a.runUse(args[1:])
	case "doctor":
		return a.runDoctor()
	case "system":
		return a.runSystem(args[1:])
	case "completion":
		return a.runCompletion(args[1:])
	case "version":
		_, err := fmt.Fprintln(a.stdout, a.version)
		return err
	case "help", "-h", "--help":
		a.printHelp()
		return nil
	default:
		return exitError{code: 2, err: fmt.Errorf("unknown command %q", args[0])}
	}
}

func (a App) runOn(args []string) error {
	cfg, err := a.loadConfig()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("on", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "")
	port := fs.Int("port", cfg.Port, "")
	if err := fs.Parse(args); err != nil {
		return exitError{code: 2, err: err}
	}
	cfg.Host = *host
	cfg.Port = *port
	cfg, shellType, shellPath, err := a.mergeShellCustomNoProxy(cfg)
	if err != nil {
		return err
	}
	noProxy, _, err := a.currentNoProxy(cfg.NoProxyCustom)
	if err != nil {
		return exitError{code: 1, err: err}
	}
	if err := a.shellManager.Enable(shellPath, shellType, shell.Env{
		Host:    cfg.Host,
		Port:    cfg.Port,
		NoProxy: noProxy,
	}); err != nil {
		return exitError{code: 1, err: err}
	}
	if err := a.saveConfig(cfg); err != nil {
		return err
	}
	return a.renderShellStatus(shellPath, shellType)
}

func (a App) runOff(args []string) error {
	cfg, err := a.loadConfig()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("off", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return exitError{code: 2, err: err}
	}
	shellType, shellPath, err := a.resolveShell(cfg)
	if err != nil {
		return err
	}
	if err := a.shellManager.Disable(shellPath); err != nil {
		return exitError{code: 1, err: err}
	}
	return a.renderShellStatus(shellPath, shellType)
}

func (a App) runStatus(args []string) error {
	cfg, err := a.loadConfig()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return exitError{code: 2, err: err}
	}
	shellType, shellPath, err := a.resolveShell(cfg)
	if err != nil {
		return err
	}
	return a.renderShellStatus(shellPath, shellType)
}

func (a App) runDetect() error {
	cfg, err := a.loadConfig()
	if err != nil {
		return err
	}
	noProxy, networks, err := a.currentNoProxy(cfg.NoProxyCustom)
	if err != nil {
		return exitError{code: 1, err: err}
	}
	renderer := output.Renderer{Color: isTerminal(a.stdout) && os.Getenv("NO_COLOR") == ""}
	return renderer.Detect(a.stdout, networks, noProxy)
}

func (a App) runSet(args []string) error {
	cfg, err := a.loadConfig()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("set", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "")
	port := fs.Int("port", cfg.Port, "")
	service := fs.String("service", cfg.NetworkService, "")
	noProxyAdd := fs.String("no-proxy-add", "", "")
	clearNoProxy := fs.Bool("no-proxy-clear-custom", false, "")
	if err := fs.Parse(args); err != nil {
		return exitError{code: 2, err: err}
	}
	changed := false
	if *host != cfg.Host || *port != cfg.Port {
		if *host == "" || *port <= 0 {
			return exitError{code: 2, err: errors.New("set requires a valid --host and --port")}
		}
		cfg.Host = *host
		cfg.Port = *port
		changed = true
	}
	if *service != "" {
		cfg.NetworkService = *service
		changed = true
	}
	if *clearNoProxy {
		cfg.NoProxyCustom = nil
		changed = true
	}
	if *noProxyAdd != "" {
		for _, item := range strings.Split(*noProxyAdd, ",") {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			if !contains(cfg.NoProxyCustom, item) {
				cfg.NoProxyCustom = append(cfg.NoProxyCustom, item)
				changed = true
			}
		}
	}
	if !changed {
		return exitError{code: 2, err: errors.New("set requires at least one change")}
	}
	if cfg.ShellType == "" {
		if shellType, detectErr := a.shellManager.Detect(os.Getenv("SHELL")); detectErr == nil {
			cfg.ShellType = string(shellType)
		}
	}
	if err := a.saveConfig(cfg); err != nil {
		return err
	}
	_, writeErr := fmt.Fprintf(a.stdout, "saved host=%s port=%d\n", cfg.Host, cfg.Port)
	return writeErr
}

func (a App) runList() error {
	services, err := a.system.ListServices()
	if err != nil {
		return exitError{code: 1, err: err}
	}
	for _, service := range services {
		fmt.Fprintln(a.stdout, service)
	}
	return nil
}

func (a App) runUse(args []string) error {
	if len(args) < 1 {
		return exitError{code: 2, err: errors.New("use requires a network service name")}
	}
	name := strings.Join(args, " ")
	return a.saveService(name)
}

func (a App) runDoctor() error {
	cfg, err := a.loadConfig()
	if err != nil {
		return err
	}
	checks := []output.Check{{Name: "platform", Status: "on", Details: "macOS supported"}}
	if path, pathErr := config.DefaultPath(); pathErr == nil {
		checks = append(checks, output.Check{Name: "config", Status: "on", Details: path})
	} else {
		checks = append(checks, output.Check{Name: "config", Status: "warn", Details: pathErr.Error()})
	}
	shellType, shellPath, shellErr := a.resolveShell(cfg)
	if shellErr != nil {
		checks = append(checks, output.Check{Name: "shell", Status: "warn", Details: shellErr.Error()})
	} else {
		checks = append(checks, output.Check{Name: "shell", Status: "on", Details: string(shellType)})
		shellStatus, statusErr := a.shellManager.Status(shellPath, shellType)
		if statusErr != nil {
			checks = append(checks, output.Check{Name: "managed", Status: "warn", Details: statusErr.Error()})
		} else if shellStatus.BlockExists {
			checks = append(checks, output.Check{Name: "managed", Status: "on", Details: shellPath})
		} else {
			checks = append(checks, output.Check{Name: "managed", Status: "warn", Details: "proxy block not present"})
		}
	}
	networks, netErr := network.DetectLocalNetworks()
	if netErr != nil {
		checks = append(checks, output.Check{Name: "networks", Status: "warn", Details: netErr.Error()})
	} else {
		checks = append(checks, output.Check{Name: "networks", Status: "on", Details: fmt.Sprintf("%d detected", len(networks))})
	}
	noProxy := network.GenerateNoProxyList(networks, cfg.NoProxyCustom)
	checks = append(checks, output.Check{Name: "no_proxy", Status: "on", Details: fmt.Sprintf("%d entries", len(noProxy))})

	target := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	if conn, dialErr := net.DialTimeout("tcp", target, 1200*time.Millisecond); dialErr != nil {
		checks = append(checks, output.Check{Name: "port", Status: "warn", Details: target + " unreachable"})
	} else {
		_ = conn.Close()
		checks = append(checks, output.Check{Name: "port", Status: "on", Details: target + " reachable"})
	}
	renderer := output.Renderer{Color: isTerminal(a.stdout) && os.Getenv("NO_COLOR") == ""}
	return renderer.Doctor(a.stdout, checks)
}

func (a App) runSystem(args []string) error {
	if len(args) == 0 {
		return exitError{code: 2, err: errors.New("system requires a subcommand: on, off, or status")}
	}
	switch args[0] {
	case "on":
		return a.runSystemOn(args[1:])
	case "off":
		return a.runSystemOff(args[1:])
	case "status":
		return a.runSystemStatus(args[1:])
	default:
		return exitError{code: 2, err: fmt.Errorf("unknown system subcommand %q", args[0])}
	}
}

func (a App) runSystemOn(args []string) error {
	cfg, err := a.loadConfig()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("system on", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	host := fs.String("host", cfg.Host, "")
	port := fs.Int("port", cfg.Port, "")
	service := fs.String("service", cfg.NetworkService, "")
	if err := fs.Parse(args); err != nil {
		return exitError{code: 2, err: err}
	}
	cfg.Host = *host
	cfg.Port = *port
	cfg, _, _, err = a.mergeShellCustomNoProxy(cfg)
	if err != nil {
		return err
	}
	resolvedService, err := a.resolveService(*service)
	if err != nil {
		return err
	}
	noProxy, _, err := a.currentNoProxy(cfg.NoProxyCustom)
	if err != nil {
		return exitError{code: 1, err: err}
	}
	if err := a.system.Enable(resolvedService, *host, *port, noProxy); err != nil {
		return exitError{code: 1, err: err}
	}
	cfg.NetworkService = resolvedService
	if err := a.saveConfig(cfg); err != nil {
		return err
	}
	return a.renderSystemStatus(config.Config{Host: *host, Port: *port, NetworkService: resolvedService})
}

func (a App) runSystemOff(args []string) error {
	cfg, err := a.loadConfig()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("system off", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	service := fs.String("service", cfg.NetworkService, "")
	if err := fs.Parse(args); err != nil {
		return exitError{code: 2, err: err}
	}
	resolvedService, err := a.resolveService(*service)
	if err != nil {
		return err
	}
	if err := a.system.Disable(resolvedService); err != nil {
		return exitError{code: 1, err: err}
	}
	return a.renderSystemStatus(config.Config{Host: cfg.Host, Port: cfg.Port, NetworkService: resolvedService})
}

func (a App) runSystemStatus(args []string) error {
	cfg, err := a.loadConfig()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("system status", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	service := fs.String("service", cfg.NetworkService, "")
	if err := fs.Parse(args); err != nil {
		return exitError{code: 2, err: err}
	}
	resolvedService, err := a.resolveService(*service)
	if err != nil {
		return err
	}
	cfg.NetworkService = resolvedService
	return a.renderSystemStatus(cfg)
}

func (a App) runCompletion(args []string) error {
	if len(args) != 1 {
		return exitError{code: 2, err: errors.New("completion requires one shell: zsh, bash, or fish")}
	}
	script, err := completionScript(args[0])
	if err != nil {
		return exitError{code: 2, err: err}
	}
	_, writeErr := io.WriteString(a.stdout, script)
	return writeErr
}

func (a App) loadConfig() (config.Config, error) {
	path, err := config.DefaultPath()
	if err != nil {
		return config.Config{}, exitError{code: 1, err: err}
	}
	cfg, err := config.Load(path)
	if err != nil {
		return config.Config{}, exitError{code: 2, err: err}
	}
	return cfg, nil
}

func (a App) saveConfig(cfg config.Config) error {
	path, err := config.DefaultPath()
	if err != nil {
		return exitError{code: 1, err: err}
	}
	if err := config.Save(path, cfg); err != nil {
		return exitError{code: 1, err: err}
	}
	return nil
}

func (a App) resolveService(current string) (string, error) {
	if current != "" {
		return current, nil
	}
	service, err := a.system.DetectService()
	if err != nil {
		return "", exitError{code: 1, err: fmt.Errorf("detect network service: %w", err)}
	}
	return service, nil
}

func (a App) resolveShell(cfg config.Config) (shell.Type, string, error) {
	if cfg.ShellType != "" {
		shellType := shell.Type(cfg.ShellType)
		path, err := a.shellManager.ConfigPath(shellType)
		if err == nil {
			return shellType, path, nil
		}
	}
	shellType, err := a.shellManager.Detect(os.Getenv("SHELL"))
	if err != nil {
		return "", "", exitError{code: 1, err: err}
	}
	path, err := a.shellManager.ConfigPath(shellType)
	if err != nil {
		return "", "", exitError{code: 1, err: err}
	}
	return shellType, path, nil
}

func (a App) currentNoProxy(custom []string) ([]string, []network.LocalNetwork, error) {
	networks, err := network.DetectLocalNetworks()
	if err != nil {
		return nil, nil, err
	}
	return network.GenerateNoProxyList(networks, custom), networks, nil
}

func (a App) existingCustomNoProxy(path string, shellType shell.Type) ([]string, error) {
	status, err := a.shellManager.Status(path, shellType)
	if err != nil {
		return nil, err
	}
	if !status.BlockExists {
		return nil, nil
	}
	existing := status.Values["NO_PROXY"]
	if existing == "" {
		existing = status.Values["no_proxy"]
	}
	if existing == "" {
		return nil, nil
	}
	networks, err := network.DetectLocalNetworks()
	if err != nil {
		return nil, err
	}
	return preserveCustomNoProxy(existing, networks, nil), nil
}

func (a App) mergeShellCustomNoProxy(cfg config.Config) (config.Config, shell.Type, string, error) {
	shellType, shellPath, err := a.resolveShell(cfg)
	if err != nil {
		return cfg, "", "", err
	}
	cfg.ShellType = string(shellType)
	existingCustom, err := a.existingCustomNoProxy(shellPath, shellType)
	if err != nil {
		return cfg, "", "", exitError{code: 1, err: err}
	}
	cfg.NoProxyCustom = mergeStrings(cfg.NoProxyCustom, existingCustom)
	return cfg, shellType, shellPath, nil
}

func (a App) renderShellStatus(path string, shellType shell.Type) error {
	status, err := a.shellManager.Status(path, shellType)
	if err != nil {
		return exitError{code: 1, err: err}
	}
	renderer := output.Renderer{Color: isTerminal(a.stdout) && os.Getenv("NO_COLOR") == ""}
	return renderer.ShellStatus(a.stdout, status)
}

func (a App) renderSystemStatus(cfg config.Config) error {
	status, err := a.system.Status(cfg.NetworkService)
	if err != nil {
		return exitError{code: 1, err: err}
	}
	renderer := output.Renderer{Color: isTerminal(a.stdout) && os.Getenv("NO_COLOR") == ""}
	return renderer.SystemStatus(a.stdout, status, cfg.Host, cfg.Port)
}

func (a App) saveService(name string) error {
	services, err := a.system.ListServices()
	if err != nil {
		return exitError{code: 1, err: err}
	}
	if !contains(services, name) {
		return exitError{code: 2, err: fmt.Errorf("network service %q not found", name)}
	}
	cfg, err := a.loadConfig()
	if err != nil {
		return err
	}
	cfg.NetworkService = name
	if err := a.saveConfig(cfg); err != nil {
		return err
	}
	_, writeErr := fmt.Fprintf(a.stdout, "saved network service %s\n", name)
	return writeErr
}

func (a App) printHelp() {
	fmt.Fprintln(a.stdout, "proxy-sw commands:")
	fmt.Fprintln(a.stdout, "  on [--host HOST] [--port PORT]")
	fmt.Fprintln(a.stdout, "  off")
	fmt.Fprintln(a.stdout, "  status")
	fmt.Fprintln(a.stdout, "  detect")
	fmt.Fprintln(a.stdout, "  doctor")
	fmt.Fprintln(a.stdout, "  set --host HOST --port PORT [--no-proxy-add VALUES] [--no-proxy-clear-custom]")
	fmt.Fprintln(a.stdout, "  list")
	fmt.Fprintln(a.stdout, "  use NAME")
	fmt.Fprintln(a.stdout, "  system on|off|status [--service NAME]")
	fmt.Fprintln(a.stdout, "  completion zsh|bash|fish")
	fmt.Fprintln(a.stdout, "  version")
}

func isTerminal(w io.Writer) bool {
	type fdWriter interface{ Fd() uintptr }
	fdw, ok := w.(fdWriter)
	if !ok {
		return false
	}
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode()&os.ModeCharDevice) != 0 && fdw.Fd() == os.Stdout.Fd()
}

func mergeStrings(parts ...[]string) []string {
	var out []string
	seen := map[string]bool{}
	for _, items := range parts {
		for _, item := range items {
			item = strings.TrimSpace(item)
			if item == "" || seen[item] {
				continue
			}
			seen[item] = true
			out = append(out, item)
		}
	}
	return out
}

func preserveCustomNoProxy(existingCSV string, networks []network.LocalNetwork, configured []string) []string {
	existingCustom := network.UserCustomNoProxy(network.ParseNoProxyCSV(existingCSV), networks)
	return mergeStrings(configured, existingCustom)
}

func completionScript(shellName string) (string, error) {
	switch shellName {
	case "bash":
		return `# bash completion for proxy-sw
_proxy_sw_completions() {
  COMPREPLY=()
  local cur="${COMP_WORDS[COMP_CWORD]}"
  local commands="on off status detect doctor set list use system completion version help"
  COMPREPLY=( $(compgen -W "${commands}" -- "${cur}") )
}
complete -F _proxy_sw_completions proxy-sw
`, nil
	case "zsh":
		return `#compdef proxy-sw
_proxy_sw() {
  local -a commands
  commands=(
    'on:enable shell proxy'
    'off:disable shell proxy'
    'status:show shell proxy status'
    'detect:show local networks and no_proxy'
    'doctor:run local diagnostics'
    'set:save host, port, and no_proxy custom items'
    'list:list macOS network services'
    'use:save default macOS network service'
    'system:manage macOS system proxy'
    'completion:print shell completion'
    'version:print version'
    'help:show help'
  )
  _describe 'command' commands
}
compdef _proxy_sw proxy-sw
`, nil
	case "fish":
		return `complete -c proxy-sw -f
complete -c proxy-sw -n '__fish_use_subcommand' -a 'on off status detect doctor set list use system completion version help'
`, nil
	default:
		return "", fmt.Errorf("unsupported shell %q", shellName)
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
