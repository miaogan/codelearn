package main

import (
	"fmt"
	"log"

	"codelearn/config"
	"codelearn/server"
)

func main() {
	cfg := config.Load()

	// 默认 JWT 密钥仅用于本地开发：生产环境务必通过 JWT_SECRET 覆盖
	if cfg.JWTSecret == "codelearn-dev-secret-change-me" {
		log.Println("⚠️  警告: 正在使用默认 JWT 密钥（codelearn-dev-secret-change-me），生产环境请设置 JWT_SECRET 环境变量")
	}

	r, err := server.New(cfg)
	if err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 CodeLearn 服务启动: http://localhost:%s", cfg.Port)
	log.Printf("   API 文档: http://localhost:%s/api/courses", cfg.Port)
	if cfg.LLMEnabled() {
		log.Printf("   AI 习题生成: 已启用 (模型: %s)", cfg.LLMModel)
	} else {
		log.Printf("   AI 习题生成: 未启用 (请设置 LLM_API_KEY 环境变量)")
	}

	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
