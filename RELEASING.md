# Releasing `proxy-sw`

## First release checklist

1. Create the GitHub repository as `kelaocai/proxy-sw`
2. Add repository secret `HOMEBREW_TAP_TOKEN`
3. Push the default branch
4. Create and push a tag like `v0.1.0`

## Commands

```bash
cd code/proxy-sw
git init
git add .
git commit -m "feat: bootstrap proxy-sw"
git branch -M main
git remote add origin git@github.com:kelaocai/proxy-sw.git
git push -u origin main
git tag v0.1.0
git push origin v0.1.0
```

## What the tag does

- `release.yml` builds macOS arm64 and amd64 binaries
- GitHub Releases receives `proxy-sw_<version>_macos_arm64.tar.gz`
- `update-homebrew-tap.yml` rewrites `packaging/homebrew/proxy-sw.rb`
- The tap repo gets updated with the new Formula and SHA256 values

## Verification

Before release, test the local install flow:

```bash
cd /Users/kelaocai/code/proxy-sw
make install-local
export PATH="$HOME/.local/bin:$PATH"
proxy-sw --help
proxy-sw set --http-host 127.0.0.1 --http-port 7897
proxy-sw detect
proxy-sw doctor
make uninstall-local
```

After the tag workflow completes:

```bash
brew tap kelaocai/tap
brew install proxy-sw
proxy-sw --help
proxy-sw set --http-host 127.0.0.1 --http-port 7897
proxy-sw detect
proxy-sw doctor
```
