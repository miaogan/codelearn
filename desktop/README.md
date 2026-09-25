# CodeLearn 桌面版

基于 [Wails v2](https://wails.io)（Go + WebView）的 CodeLearn 桌面客户端。
桌面壳内嵌启动后端（Gin + SQLite），Wails 窗口将全部请求代理到内嵌服务，
SPA 与 API 同源，前端代码与 Web 版完全共用，零改动。

## 目录结构

- `main.go`：桌面入口（内嵌后端 + 反向代理 + Wails 窗口）
- `frontend/`：前端代理目录，构建命令转发到 `../../frontend`（与 Web 版共用同一份前端源码）
- `backend/`：复用 Web 版后端（`server` 包提供可嵌入的 `server.New`）
- `build/`：Wails 平台打包资源（图标 / Info.plist / 安装器配置）

## 本地构建

依赖：Go 1.25+、Node 18+、Wails CLI（`go install github.com/wailsapp/wails/v2/cmd/wails@latest`）。
Linux 还需：`libgtk-3-dev`、`libwebkit2gtk-4.1-dev`、`pkg-config`。
`wails.json` 已内置 `"build:tags": "webkit2_41"`，构建默认使用 webkit2gtk-4.1，无需手动传参。

```bash
cd desktop
wails build
# 产物：desktop/build/bin/codelearn-desktop
```

开发模式：

```bash
cd desktop
wails dev
```

## 前端资源定位

桌面版不把前端打包进二进制，运行时按以下顺序定位构建产物
（见 `main.go` 的 `resolveFrontendDir`）：

1. 可执行文件同级 `frontend/dist/`
2. 可执行文件上一级 `frontend/dist/`（仓库开发布局：`desktop/build/bin/` → `frontend/dist/`）
3. 回退 `../frontend/dist`（相对工作目录）

分发单文件二进制时，请将 `frontend/dist` 目录放在可执行文件旁
（`<exe>/frontend/dist` 或 `<exe>/dist` 均可被识别）。

## 数据与配置

- 数据库：`~/.config/codelearn/codelearn.db`（用户配置目录，避免污染工作区）
- 内嵌服务仅监听 `127.0.0.1` 随机空闲端口，不对外暴露
- 端口冲突：每次启动自动选取空闲端口

## 跨平台打包

`wails build` 需在目标平台上执行（Windows / macOS / Linux），
打包产物请参考 Wails 官方文档与 `build/` 目录下的平台配置。
