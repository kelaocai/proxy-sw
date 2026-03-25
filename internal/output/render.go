package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/kelaocai/proxy-sw/internal/network"
	"github.com/kelaocai/proxy-sw/internal/platform/macos"
	"github.com/kelaocai/proxy-sw/internal/shell"
)

type Renderer struct {
	Color bool
}

type Check struct {
	Name    string
	Status  string
	Details string
}

func (r Renderer) SystemStatus(w io.Writer, status macos.Status, expectedHost string, expectedPort int) error {
	bypass := "-"
	if len(status.BypassDomains) > 0 {
		bypass = strings.Join(status.BypassDomains, ",")
	}
	_, err := fmt.Fprintf(w, "%s\n%s\n%s\n%s\n%s\n%s\n",
		r.header("System Proxy", status.NetworkService),
		r.systemLine("web", status.Web),
		r.systemLine("https", status.HTTPS),
		r.systemLine("socks", status.SOCKS),
		r.kvLine("bypass", bypass),
		r.endpoint(expectedHost, expectedPort),
	)
	return err
}

func (r Renderer) ShellStatus(w io.Writer, status shell.Status) error {
	blockState := "off"
	if status.BlockExists {
		blockState = "on"
	}
	_, err := fmt.Fprintf(w, "%s\n%s\n%s\n%s\n%s\n",
		r.header("Shell Proxy", string(status.ShellType)),
		fmt.Sprintf("%s %-12s %s", r.dot("on"), "file", status.Path),
		fmt.Sprintf("%s %-12s %s", r.dot(blockState), "managed", blockState),
		r.kvLine("http_proxy", status.Values["http_proxy"]),
		r.kvLine("no_proxy", status.Values["no_proxy"]),
	)
	return err
}

func (r Renderer) Detect(w io.Writer, networks []network.LocalNetwork, noProxy []string) error {
	if _, err := fmt.Fprintln(w, "Local Networks"); err != nil {
		return err
	}
	for _, item := range networks {
		if _, err := fmt.Fprintf(w, "%-8s %-15s %s\n", item.Interface, item.IPAddress, item.NetworkCIDR); err != nil {
			return err
		}
	}
	if len(networks) == 0 {
		if _, err := fmt.Fprintln(w, "none"); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "\nno_proxy %s\n", strings.Join(noProxy, ","))
	return err
}

func (r Renderer) Doctor(w io.Writer, checks []Check) error {
	if _, err := fmt.Fprintln(w, "Proxy Doctor"); err != nil {
		return err
	}
	for _, check := range checks {
		if _, err := fmt.Fprintf(w, "%s %-12s %s\n", r.dot(check.Status), check.Name, check.Details); err != nil {
			return err
		}
	}
	return nil
}

func (r Renderer) header(title, suffix string) string {
	if suffix == "" {
		return title
	}
	return title + "  " + suffix
}

func (r Renderer) systemLine(label string, state macos.ProxyState) string {
	indicator := r.dot("off")
	statusText := "off"
	target := "-"
	if !state.Available {
		indicator = r.dot("warn")
		statusText = "unavailable"
	} else if state.Enabled {
		indicator = r.dot("on")
		statusText = "on"
		target = fmt.Sprintf("%s:%d", state.Server, state.Port)
	}
	return fmt.Sprintf("%s %-6s %-11s %s", indicator, label, statusText, target)
}

func (r Renderer) endpoint(host string, port int) string {
	return fmt.Sprintf("  endpoint %-11s %s:%d", "", host, port)
}

func (r Renderer) kvLine(key, value string) string {
	state := "warn"
	details := value
	if value == "" {
		details = "-"
	} else {
		state = "on"
	}
	return fmt.Sprintf("%s %-12s %s", r.dot(state), key, details)
}

func (r Renderer) dot(kind string) string {
	plain := "●"
	if !r.Color {
		return plain
	}
	switch kind {
	case "on":
		return "\033[32m●\033[0m"
	case "warn":
		return "\033[33m●\033[0m"
	default:
		return "\033[31m●\033[0m"
	}
}
