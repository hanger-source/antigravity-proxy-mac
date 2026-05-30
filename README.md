# Antigravity Proxy (macOS)

macOS 菜单栏工具，通过 TUN + 进程/域名路由规则，**只代理指定 App 的流量**，其余流量直连。

## 原理

```
所有流量 → TUN 虚拟网卡 → sing-box 路由判断
  ├── process_name 匹配（Codex/Antigravity/...）→ 走代理
  ├── domain_suffix 匹配（openai.com/...）→ 走代理
  └── 其余所有进程 → 直连（不受影响）
```

与全局 TUN 代理的区别：`route.final = "direct"`，只有匹配的进程或域名走代理。

## 典型用途

- **Codex App 登录**：Codex 的 OAuth token exchange 不走系统代理，需要 TUN 强制代理
- **Antigravity IDE**：IDE 相关进程走代理
- **任何需要翻墙的 GUI App**：配置进程名即可

## 安装

```bash
# 编译
make build

# 安装到 /Applications
make install
```

首次启动会弹出系统授权框安装 helper（仅一次）。

## 配置

配置文件：`~/.antigravity-proxy/config.json`

首次启动自动生成默认配置。

### 上游代理模式（推荐）

指向本地已有的代理（V2RayX、Clash 等）：

```json
{
  "upstream": {
    "type": "socks5",
    "host": "127.0.0.1",
    "port": 13658
  },
  "target_processes": [
    "Codex",
    "Codex Helper",
    "Codex Helper (Renderer)",
    "Codex Helper (GPU)",
    "Codex Helper (Plugin)",
    "Antigravity",
    "Antigravity Helper",
    "Antigravity Helper (Renderer)"
  ],
  "target_domains": [
    "openai.com",
    "auth.openai.com",
    "api.openai.com",
    "chatgpt.com",
    "oaistatic.com",
    "oaiusercontent.com"
  ],
  "log_level": "info"
}
```

### 直连节点模式

不依赖本地代理，直接配置节点：

```json
{
  "nodes": [
    {
      "name": "我的节点",
      "type": "vmess",
      "server": "example.com",
      "port": 443,
      "uuid": "xxx-xxx-xxx"
    }
  ],
  "selected_node": 0,
  "target_processes": ["Codex", "Codex Helper", "Codex Helper (Renderer)"],
  "target_domains": ["openai.com"]
}
```

### 配置项

| 字段 | 说明 |
|---|---|
| `upstream.type` | 上游代理类型：`socks5` 或 `http` |
| `upstream.host` | 上游代理地址（通常 `127.0.0.1`） |
| `upstream.port` | 上游代理端口 |
| `target_processes` | 要代理的进程名列表 |
| `target_domains` | 要代理的域名后缀列表（匹配任何进程） |
| `nodes` | 直连节点列表（upstream 优先） |
| `log_level` | sing-box 日志级别：trace/debug/info/warn/error |

### 路由优先级

1. 私有 IP → 直连
2. `target_domains` 匹配 → 走代理（不管哪个进程）
3. `target_processes` 匹配 → 走代理（不管访问什么域名）
4. 其余 → 直连

## 如何找到 App 的进程名

```bash
# 方法 1：Activity Monitor 里看
# 方法 2：命令行
ps aux | grep -i "codex" | grep -v grep
```

常见 App 进程名：

| App | 进程名 |
|---|---|
| Codex | `Codex`, `Codex Helper`, `Codex Helper (Renderer)`, `Codex Helper (GPU)`, `Codex Helper (Plugin)` |
| Antigravity | `Antigravity`, `Antigravity Helper`, ... |
| Claude | `Claude`, `Claude Helper`, `Claude Helper (Renderer)` |
| Cursor | `Cursor`, `Cursor Helper`, `Cursor Helper (Renderer)` |

## 日志

- 应用日志：`~/.antigravity-proxy/antigravity-proxy.log`
- sing-box 日志：`~/.antigravity-proxy/singbox.log`
- Helper 日志：`/var/log/antigravity-proxy-helper.log`

## 依赖

- sing-box 1.11.x（首次启动自动下载）
- macOS Security.framework（helper 安装授权）
