# proxy-sw 改进需求：自动添加本地地址不走代理

## 经验总结

### 本次配置过程中遇到的问题

1. **环境变量分散管理**
   - 系统级代理（networksetup）和终端环境变量（http_proxy）需要分别配置
   - 各工具（git、npm、yarn、pnpm）都有自己的代理配置，需要逐一设置
   - 没有统一的工具来管理所有这些配置

2. **no_proxy 设置容易被忽略**
   - 大部分代理配置教程只关注如何设置代理，忽略了本地网段绕过
   - 默认情况下，本地开发服务器、内网服务都会走代理，导致访问缓慢或失败
   - 需要手动计算并配置本地网段（如 192.168.2.0/24）

3. **配置持久化困难**
   - 终端环境变量需要添加到 ~/.zshrc 才能持久化
   - 不同工具的代理配置分散在不同位置
   - 开启/关闭代理需要修改多处配置

### 理想的用户体验

```bash
# 一键开启代理（系统 + 终端 + 所有工具）
proxy-sw on --shell

# 自动检测并配置本地网段绕过
proxy-sw on --detect-local-network

# 查看当前所有代理配置状态
proxy-sw status --all

# 一键关闭所有代理
proxy-sw off --all
```

---

## 改进需求

### 1. 新增功能：终端环境变量管理

**目标**：在开启/关闭系统代理的同时，自动管理终端环境变量

**实现方案**：

```go
// internal/shell/shell.go
package shell

type Manager interface {
    // 将环境变量配置写入 shell 配置文件
    Enable(host string, port int, noProxy []string) error
    // 从 shell 配置文件移除环境变量配置
    Disable() error
    // 获取当前 shell 类型
    DetectShell() (ShellType, error)
}

type ShellType string
const (
    Zsh  ShellType = "zsh"
    Bash ShellType = "bash"
    Fish ShellType = "fish"
)
```

**配置写入位置**：
- Zsh: `~/.zshrc`
- Bash: `~/.bashrc` 或 `~/.bash_profile`
- Fish: `~/.config/fish/config.fish`

**配置格式**（使用标记便于管理）：
```bash
# === proxy-sw managed start ===
export http_proxy="http://127.0.0.1:7897"
export https_proxy="http://127.0.0.1:7897"
export all_proxy="socks5://127.0.0.1:7897"
export HTTP_PROXY="http://127.0.0.1:7897"
export HTTPS_PROXY="http://127.0.0.1:7897"
export ALL_PROXY="socks5://127.0.0.1:7897"
export no_proxy="localhost,127.0.0.1,::1,192.168.0.0/16,10.0.0.0/8,172.16.0.0/12,192.168.2.0/24,*.local"
export NO_PROXY="localhost,127.0.0.1,::1,192.168.0.0/16,10.0.0.0/8,172.16.0.0/12,192.168.2.0/24,*.local"
# === proxy-sw managed end ===
```

**CLI 接口**：
```bash
# 开启系统代理 + 终端环境变量
proxy-sw on --shell

# 仅开启终端环境变量（不修改系统代理）
proxy-sw on --shell-only

# 查看终端代理状态
proxy-sw status --shell
```

---

### 2. 新增功能：自动检测本地网段

**目标**：自动检测当前连接的本地网络，生成 no_proxy 绕过规则

**实现方案**：

```go
// internal/network/detect.go
package network

// LocalNetwork 表示本地网络信息
type LocalNetwork struct {
    Interface   string   // 网卡名称，如 "en0"
    IPAddress   string   // IP 地址，如 "192.168.2.10"
    NetworkCIDR string   // 网段，如 "192.168.2.0/24"
    IsPrivate   bool     // 是否为私有地址
}

// DetectLocalNetworks 检测所有本地网络
// 使用 macOS 的 ifconfig 或 networksetup 获取网络信息
func DetectLocalNetworks() ([]LocalNetwork, error)

// GenerateNoProxyList 生成 no_proxy 列表
// 包含：
// - 标准私有网段：10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16
// - 当前连接的具体网段
// - localhost, 127.0.0.1, ::1
// - *.local
func GenerateNoProxyList(networks []LocalNetwork) []string
```

**CLI 接口**：
```bash
# 自动检测并配置本地网段
proxy-sw on --detect-local

# 查看检测到的本地网络
proxy-sw detect
# 输出示例：
# Interface    IP Address      Network
# en0          192.168.2.10    192.168.2.0/24
# en1          10.0.1.5        10.0.1.0/24

# 配置文件中保存检测到的网段
proxy-sw set --no-proxy-auto
```

**配置扩展**：
```yaml
# ~/.config/proxy-sw/config.yaml
host: 127.0.0.1
port: 7897
network_service: Wi-Fi

# 新增：no_proxy 配置
no_proxy:
  auto_detect: true  # 启动时自动检测本地网络
  custom:            # 自定义绕过规则
    - "*.company.com"
    - "internal.example.com"
  exclude:           # 从自动检测中排除的网段
    - "192.168.100.0/24"
```

---

### 3. 新增功能：工具链代理统一管理

**目标**：统一管理 git、npm、yarn、pnpm、pip、curl 等工具的代理配置

**实现方案**：

```go
// internal/tools/tools.go
package tools

type Tool string
const (
    Git   Tool = "git"
    Npm   Tool = "npm"
    Yarn  Tool = "yarn"
    Pnpm  Tool = "pnpm"
    Pip   Tool = "pip"
)

type ToolManager interface {
    // 设置工具的代理配置
    SetProxy(host string, port int, noProxy []string) error
    // 清除工具的代理配置
    ClearProxy() error
    // 检查工具是否已安装
    IsInstalled() bool
    // 获取当前配置
    GetCurrentProxy() (ProxyConfig, error)
}

// 批量管理所有工具
func SetAllToolsProxy(host string, port int, noProxy []string) map[Tool]error
func ClearAllToolsProxy() map[Tool]error
```

**CLI 接口**：
```bash
# 开启代理时同时配置所有工具
proxy-sw on --tools

# 单独管理工具代理
proxy-sw tools on   # 配置所有工具代理
proxy-sw tools off  # 清除所有工具代理
proxy-sw tools status  # 查看各工具代理状态

# 指定工具
proxy-sw tools on --only git,npm
```

---

### 4. 完整的工作流示例

#### 场景 1：日常开发
```bash
# 早上启动，一键开启所有代理
$ proxy-sw on --shell --detect-local --tools

# 系统代理已开启
# 终端环境变量已配置
# 本地网段 192.168.2.0/24 已添加到 no_proxy
# git/npm/yarn/pnpm 代理已配置

# 晚上关闭
$ proxy-sw off --all
```

#### 场景 2：切换到公司网络
```bash
# 公司网络有内网服务，需要更新 no_proxy
$ proxy-sw detect
检测到新网段: 10.10.0.0/16

$ proxy-sw on --detect-local
已更新 no_proxy，新增: 10.10.0.0/16
```

#### 场景 3：查看完整状态
```bash
$ proxy-sw status --all

System Proxy (Wi-Fi)
  web:    on  127.0.0.1:7897
  secure: on  127.0.0.1:7897
  socks:  on  127.0.0.1:7897

Shell Environment
  http_proxy:  http://127.0.0.1:7897
  https_proxy: http://127.0.0.1:7897
  no_proxy:    localhost,127.0.0.1,192.168.2.0/24,10.0.0.0/8,...
  Config file: ~/.zshrc

Local Networks
  en0  192.168.2.10  192.168.2.0/24
  en1  10.0.1.5      10.0.1.0/24

Toolchain
  git:   http://127.0.0.1:7897
  npm:   http://127.0.0.1:7897 (noproxy configured)
  yarn:  http://127.0.0.1:7897 (noproxy configured)
  pnpm:  not configured
```

---

### 5. 配置文件更新

```yaml
# ~/.config/proxy-sw/config.yaml
version: "2"

proxy:
  host: 127.0.0.1
  port: 7897
  network_service: Wi-Fi

# 新增：shell 环境变量管理
shell:
  enabled: true
  shell_type: auto  # auto, zsh, bash, fish
  config_file: ""   # 留空则使用默认路径

# 新增：no_proxy 配置
no_proxy:
  auto_detect: true
  standard_private_networks: true  # 10/8, 172.16/12, 192.168/16
  custom:
    - "*.company.com"
    - "internal.example.com"

# 新增：工具链配置
tools:
  auto_configure: true
  enabled:
    - git
    - npm
    - yarn
    - pnpm
```

---

### 6. 向后兼容

- 所有新增功能默认为关闭状态，不影响现有用户
- `--shell`、`--detect-local`、`--tools` 等 flag 需要显式指定
- 配置文件新增字段有合理的默认值

---

## 优先级建议

| 优先级 | 功能 | 理由 |
|--------|------|------|
| P0 | 本地网段自动检测 (no_proxy) | 解决最痛点的问题，影响日常使用 |
| P1 | 终端环境变量管理 | 与现有系统代理功能互补 |
| P2 | 工具链统一管理 | 提升开发体验，但各工具已有配置方式 |
| P3 | 配置文件格式升级 | 为更多功能预留扩展空间 |