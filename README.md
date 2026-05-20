# Antigravity Proxy (macOS)

macOS 菜单栏工具，通过 TUN + 进程路由规则，**只代理 Antigravity 相关进程**，其余流量直连。

## 原理

```
所有流量 → TUN 虚拟网卡 → sing-box 路由判断
  ├── process_name 匹配 Antigravity → 走代理
  └── 其余所有进程 → 直连（不受影响）
```

与全局 TUN 代理的区别：`route.final = "direct"`，只有 Antigravity 进程的流量走代理。

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

```json
{
  "proxy": {
    "host": "127.0.0.1",
    "port": 7890,
    "type": "socks5"
  }
}
```

首次启动自动生成默认配置（指向本地 Clash 7890 端口）。

### 配置项

| 字段 | 说明 |
|---|---|
| `proxy.host` | 本地代理地址 |
| `proxy.port` | 本地代理端口 |
| `proxy.type` | `socks5` 或 `http` |
| `target_processes` | 自定义要代理的进程名列表（默认已包含 Antigravity 全家桶） |
| `log_level` | sing-box 日志级别：trace/debug/info/warn/error |

### 直连节点模式

如果不想依赖本地 Clash，可以直接配置节点：

```json
{
  "proxy": {},
  "nodes": [
    {
      "name": "我的节点",
      "type": "vmess",
      "server": "example.com",
      "port": 443,
      "uuid": "xxx-xxx-xxx"
    }
  ],
  "selected_node": 0
}
```

## 日志

- 应用日志：`~/.antigravity-proxy/antigravity-proxy.log`
- sing-box 日志：`~/.antigravity-proxy/singbox.log`
- Helper 日志：`/var/log/antigravity-proxy-helper.log`

## 默认代理的进程

| 进程名 | 来源 |
|---|---|
| Antigravity | Antigravity.app 主进程 |
| Antigravity Helper | Antigravity.app 子进程 |
| Antigravity Helper (Renderer) | 渲染进程 |
| language_server | AI 语言服务器 |
| Electron | Antigravity IDE 主进程 |
| Antigravity IDE Helper | IDE 子进程 |
| Antigravity IDE Helper (Plugin) | 插件宿主 |
| Antigravity IDE Helper (Renderer) | IDE 渲染 |
| Antigravity IDE Helper (GPU) | IDE GPU |
| language_server_macos_arm | IDE AI 语言服务器 |

## 依赖

- sing-box 1.11.x（首次启动自动下载）
- macOS Security.framework（helper 安装授权）
