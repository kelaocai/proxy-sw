# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Changed
- Repositioned `proxy-sw` as a shell-first proxy environment manager
- Moved macOS system proxy controls under `proxy-sw system ...`

### Added
- Automatic local-network detection for `no_proxy`
- Shell managed block support for `zsh`, `bash`, and `fish`
- `detect` command for local network and `no_proxy` preview

## [0.1.0] - 2026-03-09

### Added
- Initial macOS-first `proxy-sw` CLI
- System proxy `on`, `off`, and `status` commands
- Config persistence in `~/.config/proxy-sw/config.yaml`
- Network service discovery, listing, and selection
- Shell completion output for `zsh`, `bash`, and `fish`
- GitHub Actions workflows for CI, release, and Homebrew tap sync
- Homebrew Formula templates for source repo and tap repo
