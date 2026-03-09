# proxy-sw

`proxy-sw` is a macOS-first CLI for managing terminal proxy environment variables for local proxy tools like Clash.

## Scope

Version 1 now focuses on shell proxy environment management with automatic `no_proxy` generation.

- Manages shell `http_proxy`, `https_proxy`, `all_proxy`, and `no_proxy`
- Automatically detects local private networks and adds them to `no_proxy`
- Stores defaults in `~/.config/proxy-sw/config.yaml`
- Ships as a single Go binary
- Designed for Homebrew tap installation
- Keeps macOS system proxy support under `proxy-sw system ...` for advanced use

Out of scope for v1:

- Starting or checking Clash itself
- Linux or Windows proxy management
- Tool-specific proxy config for git, npm, pip, and similar tools

## Install

### Homebrew tap

```bash
brew tap kelaocai/tap
brew install proxy-sw
```

Until the first tagged release is published, build locally:

```bash
git clone https://github.com/kelaocai/proxy-sw.git
cd proxy-sw
make build
```

For a local user-style install without Homebrew:

```bash
make install-local
export PATH="$HOME/.local/bin:$PATH"
proxy-sw --help
```

## Commands

```bash
proxy-sw set --host 127.0.0.1 --port 7897
proxy-sw on
proxy-sw detect
proxy-sw doctor
proxy-sw status
proxy-sw off
```

## Quick Start

For most users, this is the full setup flow:

```bash
proxy-sw set --host 127.0.0.1 --port 7897
proxy-sw on
proxy-sw detect
proxy-sw doctor
```

Turn it off later with:

```bash
proxy-sw off
```

Advanced usage:

- `on` automatically regenerates `no_proxy` from current local private networks
- `set --no-proxy-add a,b` adds custom `no_proxy` entries
- `set --no-proxy-clear-custom` clears custom `no_proxy` entries
- `proxy-sw system on|off|status` manages macOS system proxy when you explicitly need it
- `proxy-sw list` and `proxy-sw use "Wi-Fi"` are only needed for `proxy-sw system ...`

## Config

Default config path:

```text
~/.config/proxy-sw/config.yaml
```

Example:

```yaml
host: 127.0.0.1
port: 7897
shell_type: zsh
no_proxy_custom:
  - internal.example.com
  - "*.corp.local"
network_service: Wi-Fi
```

## Release

- Tag a release like `v0.1.0`
- GitHub Actions builds macOS arm64 and amd64 archives
- The release workflow publishes assets to GitHub Releases
- The tap sync workflow updates `packaging/homebrew/proxy-sw.rb` into the tap repo
