# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.1.4] - 2026-04-11

### Fixed
- Keep the bundled Homebrew formula in the source repository in sync during tagged releases
- Allow `all_proxy` to use a dedicated `socks5` endpoint instead of always reusing the HTTP port

### Changed
- Document `brew update && brew upgrade proxy-sw` for upgrading existing Homebrew installs
- Store proxy endpoints as `http` and `socks5` objects in config, while continuing to read legacy top-level `host` and `port`

## [0.1.3] - 2026-03-25

### Fixed
- Sync generated `no_proxy` entries into macOS proxy bypass domains when running `proxy-sw system on`
- Preserve custom `no_proxy` entries recovered from the managed shell block when `proxy-sw system on` regenerates system bypass domains

### Added
- Show macOS proxy bypass domains in `proxy-sw system status`

## [0.1.2] - 2026-03-21

### Fixed
- Preserve user-added `NO_PROXY` entries when `proxy-sw on` rewrites the managed shell block

### Added
- Regression coverage for parsing and preserving custom `no_proxy` values

## [0.1.1] - 2026-03-09

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
