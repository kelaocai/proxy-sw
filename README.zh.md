# proxy-sw

[English](README.md) | 中文

`proxy-sw` 是一个面向 macOS 的命令行工具，用来管理终端里的代理环境变量，适合本地代理工具如 Clash、Clash Verge、ClashX、mihomo 等场景。

## 项目定位

当前 v1 版本聚焦在 shell 代理环境管理：

- 管理 shell 中的 `http_proxy`、`https_proxy`、`all_proxy` 和 `no_proxy`
- 自动探测本地私有网段，并加入 `no_proxy`
- 默认配置保存在 `~/.config/proxy-sw/config.yaml`
- 单文件 Go 二进制，安装简单
- 支持通过 Homebrew tap 分发
- 保留 `proxy-sw system ...` 用于需要时管理 macOS 系统代理

当前不做：

- 不负责启动或检测 Clash 本体
- 不负责 Linux / Windows 代理管理
- 不负责为 git、npm、pip 等工具单独写代理配置

## 安装

### Homebrew 安装

```bash
brew tap kelaocai/tap
brew install proxy-sw
```

如果本地已经安装过旧版本，升级要用：

```bash
brew update
brew upgrade proxy-sw
```

安装后，`proxy-sw` 默认会从这里读取和保存配置：

```text
~/.config/proxy-sw/config.yaml
```

如果你想从源码本地构建：

```bash
git clone https://github.com/kelaocai/proxy-sw.git
cd proxy-sw
make build
```

如果你不走 Homebrew，也可以本地安装到用户目录：

```bash
make install-local
export PATH="$HOME/.local/bin:$PATH"
proxy-sw --help
```

## 常用命令

```bash
proxy-sw set --host 127.0.0.1 --port 7897
proxy-sw on
proxy-sw detect
proxy-sw doctor
proxy-sw status
proxy-sw off
```

## 快速开始

大多数用户只需要这几步：

```bash
proxy-sw set --host 127.0.0.1 --port 7897
proxy-sw on
proxy-sw detect
proxy-sw doctor
```

说明：

- `set` 会把默认代理地址保存到 `~/.config/proxy-sw/config.yaml`
- `on` 会把代理环境变量写入当前 shell 的配置文件，例如 `~/.zshrc`
- `detect` 会探测本地网络并生成合适的 `no_proxy`
- `doctor` 会做基础诊断，帮助确认当前配置是否可用

如果要关闭 shell 代理：

```bash
proxy-sw off
```

## 配置文件

`set`、`on`、`status`、`detect`、`doctor`、`system` 等命令默认使用这份配置：

```text
~/.config/proxy-sw/config.yaml
```

示例：

```yaml
host: 127.0.0.1
port: 7897
shell_type: zsh
no_proxy_custom:
  - internal.example.com
  - "*.corp.local"
network_service: Wi-Fi
```

字段说明：

- `host`：代理主机地址
- `port`：代理端口
- `shell_type`：当前 shell 类型，例如 `zsh`
- `no_proxy_custom`：额外追加的 `no_proxy` 规则
- `network_service`：系统代理使用的默认网络服务，例如 `Wi-Fi`

## Shell 配置写入位置

`proxy-sw on` 和 `proxy-sw off` 不只是读取 `config.yaml`，还会修改 shell 配置文件中的托管区块：

- zsh: `~/.zshrc`
- bash: `~/.bashrc`
- fish: `~/.config/fish/config.fish`

`off` 会直接移除 `proxy-sw` 托管的配置区块，不是注释掉，而是删除。

## 进阶说明

- `on` 会根据当前本地网络自动重新生成 `no_proxy`
- `set --no-proxy-add a,b` 可追加自定义 `no_proxy` 规则
- `set --no-proxy-clear-custom` 可清空自定义 `no_proxy`
- `proxy-sw system on|off|status` 用于显式管理 macOS 系统代理，`system on` 也会把生成出的 `no_proxy` 同步到系统 bypass domains
- `proxy-sw system status` 会显示当前 macOS 系统里的 bypass domains
- `proxy-sw list` 和 `proxy-sw use "Wi-Fi"` 主要用于 `proxy-sw system ...`

## 发布

- push `v*` tag 后会触发 GitHub Actions
- `release.yml` 会构建 macOS arm64 / x86_64 二进制并发布到 GitHub Release
- `update-homebrew-tap.yml` 会更新 Homebrew tap 中的 `Formula/proxy-sw.rb`
