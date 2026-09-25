package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"

	"codelearn/config"
	"codelearn/server"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// resolveFrontendDir 定位前端构建产物：优先可执行文件旁，其次仓库相对路径
func resolveFrontendDir() string {
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		candidates := []string{
			filepath.Join(exeDir, "frontend", "dist"),
			filepath.Join(exeDir, "dist"),
			filepath.Join(exeDir, "..", "frontend", "dist"),
		}
		for _, c := range candidates {
			if fi, err := os.Stat(filepath.Join(c, "index.html")); err == nil && !fi.IsDir() {
				return c
			}
		}
	}
	return "../frontend/dist"
}

func main() {
	cfg := config.Load()

	// 绑定本地空闲端口，避免与已运行实例冲突
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("分配本地端口失败: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()
	cfg.Port = fmt.Sprintf("%d", port)

	// 桌面版数据文件放在用户配置目录下，避免污染工作目录
	dataDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("获取配置目录失败: %v", err)
	}
	dataDir = filepath.Join(dataDir, "codelearn")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}
	cfg.DBPath = filepath.Join(dataDir, "codelearn.db")
	cfg.FrontendDir = resolveFrontendDir()

	if cfg.JWTSecret == "codelearn-dev-secret-change-me" {
		log.Println("⚠️  警告: 正在使用默认 JWT 密钥（codelearn-dev-secret-change-me），生产环境请设置 JWT_SECRET 环境变量")
	}

	// 内嵌启动后端（Gin + SQLite），仅监听 127.0.0.1
	r, err := server.New(cfg)
	if err != nil {
		log.Fatalf("启动后端失败: %v", err)
	}
	backend := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: r,
	}
	go func() {
		if err := backend.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("内嵌后端异常退出: %v", err)
		}
	}()

	// Wails 窗口：AssetServer 将全部请求代理到内嵌后端，
	// 窗口内 SPA 与 API 同源，现有前端无需任何改动。
	backendURL, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))
	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	err = wails.Run(&options.App{
		Title:     "CodeLearn",
		Width:     1200,
		Height:    800,
		MinWidth:  900,
		MinHeight: 640,
		AssetServer: &assetserver.Options{
			Assets:  nil,
			Handler: proxy,
		},
		OnStartup: func(ctx context.Context) {
			log.Printf("CodeLearn 桌面版已启动（内嵌服务 127.0.0.1:%d，数据目录 %s）", port, dataDir)
		},
		OnShutdown: func(ctx context.Context) {
			_ = backend.Shutdown(ctx)
		},
	})
	if err != nil {
		log.Fatalf("桌面应用启动失败: %v", err)
	}
}
