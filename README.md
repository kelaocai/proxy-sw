# proxy-sw

English | [中文](README.zh.md)

`proxy-sw` is a macOS-first CLI for managing terminal proxy environment variables for local proxy tools like Clash.

## Scope

Version 1 now focuses on shell proxy environment management with automatic `no_proxy` generation.

- Manages shell `http_proxy`, `https_proxy`, `all_proxy`, and `no_proxy`
- Automatically detects local private networks and adds them to `no_proxy`
- Stores defaults in `~/.config/proxy-sw/config.yaml`
- Ships as a single Go binary
- Designed for Homebrew tap installation
- Keeps macOS system proxy support under `proxy-sw system ...` for advanced use, including syncing generated `no_proxy` values to macOS proxy bypass domains on `system on`

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

When upgrading from an older installed version, run:

```bash
brew update
brew upgrade proxy-sw
```

After install, `proxy-sw` reads and writes its default config at:

```text
~/.config/proxy-sw/config.yaml
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
proxy-sw set --http-host 127.0.0.1 --http-port 7897
proxy-sw on
proxy-sw detect
proxy-sw doctor
proxy-sw status
proxy-sw off
```

## Quick Start

For most users, this is the full setup flow:

```bash
proxy-sw set --http-host 127.0.0.1 --http-port 7897
proxy-sw on
proxy-sw detect
proxy-sw doctor
```

This saves your proxy defaults to `~/.config/proxy-sw/config.yaml`, and `proxy-sw on` writes the managed proxy block into your shell config file such as `~/.zshrc`.
`http_proxy` and `https_proxy` use the `http` endpoint, while `all_proxy` uses the optional `socks5` endpoint and falls back to `http` if `socks5` is not configured.

Turn it off later with:

```bash
proxy-sw off
```

Advanced usage:

- `on` automatically regenerates `no_proxy` from current local private networks
- `set --no-proxy-add a,b` adds custom `no_proxy` entries
- `set --no-proxy-clear-custom` clears custom `no_proxy` entries
- `proxy-sw system on|off|status` manages macOS system proxy when you explicitly need it, and `system on` syncs the generated `no_proxy` list into macOS proxy bypass domains
- `proxy-sw system status` shows the current macOS bypass domain list as `bypass`
- `proxy-sw list` and `proxy-sw use "Wi-Fi"` are only needed for `proxy-sw system ...`

## Config

Default config path used by `set`, `on`, `status`, `detect`, `doctor`, and `system`:

```text
~/.config/proxy-sw/config.yaml
```

Example:

```yaml
http:
  host: 127.0.0.1
  port: 6152
socks5:
  host: 127.0.0.1
  port: 6153
shell_type: zsh
no_proxy_custom:
  - internal.example.com
  - "*.corp.local"
network_service: Wi-Fi
```

Legacy top-level `host` and `port` are still read as the HTTP endpoint for backwards compatibility. New saves use the nested structure above.

## Release

- Tag a release like `v0.1.0`
- GitHub Actions builds macOS arm64 and amd64 archives
- The release workflow publishes assets to GitHub Releases
- The tap sync workflow updates `packaging/homebrew/proxy-sw.rb` into the tap repo
