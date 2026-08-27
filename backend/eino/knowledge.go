package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"codelearn/config"
	"codelearn/repository"

	"github.com/cloudwego/eino/schema"
	openaillm "github.com/cloudwego/eino-ext/components/model/openai"
)

// KnowledgeRAG 使用 Eino 的 Retriever + Chain 模式
//
// RAG 流程：
// 1. Retriever: 根据学生问题检索相关课程内容（keyword-based）
// 2. ChatTemplate: 组装检索到的上下文 + 学生问题
// 3. ChatModel: 基于上下文生成个性化回答
//
// 这展示了 Eino RAG 的核心思想：检索增强生成
type KnowledgeRAG struct {
	cfg  *config.Config
	repo *repository.Repository
}

type KnowledgeAnswer struct {
	Answer    string            `json:"answer"`
	Sources   []KnowledgeSource `json:"sources"`
	FollowUp  []string          `json:"follow_ups"`
}

type KnowledgeSource struct {
	CourseName string `json:"course_name"`
	LessonTitle string `json:"lesson_title"`
	Snippet     string `json:"snippet"`
}

func NewKnowledgeRAG(cfg *config.Config, repo *repository.Repository) *KnowledgeRAG {
	return &KnowledgeRAG{cfg: cfg, repo: repo}
}

// Retrieve 检索相关课程内容（Eino Retriever 接口的实现）
func (k *KnowledgeRAG) Retrieve(query string, limit int) []KnowledgeSource {
	log.Printf("[知识RAG] Retriever: 检索 query=%s", truncate(query, 50))

	// 从课程内容中检索匹配的章节
	results := k.repo.SearchLessonsByKeyword(query, limit)
	sources := make([]KnowledgeSource, 0, len(results))

	for _, l := range results {
		snippet := l.Content
		if len(snippet) > 300 {
			snippet = snippet[:300] + "..."
		}
		sources = append(sources, KnowledgeSource{
			CourseName:  l.CourseName,
			LessonTitle: l.Title,
			Snippet:     snippet,
		})
	}

	log.Printf("[知识RAG] 检索到 %d 条相关内容", len(sources))
	return sources
}

// Ask 完整 RAG 链：Retrieve → ChatTemplate → Generate
func (k *KnowledgeRAG) Ask(ctx context.Context, query, language string) (*KnowledgeAnswer, error) {
	if k.cfg.LLMAPIKey == "" {
		return nil, fmt.Errorf("LLM_API_KEY 未设置")
	}

	// RAG Step1: 检索相关课程内容
	sources := k.Retrieve(query, 5)

	// RAG Step2: 组装上下文
	contextText := ""
	if len(sources) > 0 {
		var parts []string
		for i, s := range sources {
			parts = append(parts, fmt.Sprintf("[%d] 课程: %s | 章节: %s\n%s", i+1, s.CourseName, s.LessonTitle, s.Snippet))
		}
		contextText = strings.Join(parts, "\n\n---\n\n")
	} else {
		contextText = "（未找到直接相关的课程内容）"
	}

	log.Printf("[知识RAG] Step2: 组装上下文, 长度=%d", len(contextText))

	// RAG Step3: LLM 基于上下文生成回答
	chatModel, err := openaillm.NewChatModel(ctx, &openaillm.ChatModelConfig{
		APIKey:  k.cfg.LLMAPIKey,
		BaseURL: k.cfg.LLMBaseURL,
		Model:   k.cfg.LLMModel,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	messages := []*schema.Message{
		{Role: schema.System, Content: fmt.Sprintf(`你是一名 %s 编程教育助手。请根据提供的课程内容上下文回答学生的问题。

要求：
1. 优先基于课程内容回答
2. 如果课程内容中有相关知识点，引用并展开解释
3. 如果课程内容中没有直接答案，用你的知识补充，但要说明
4. 回答要简洁易懂，适合初学者
5. 提供 2-3 个相关的后续问题供学生继续学习
6. 严格返回 JSON 格式`, language)},
		{Role: schema.User, Content: fmt.Sprintf(`课程内容上下文：
%s

学生的问题：%s

请基于上下文回答，返回 JSON：
{
  "answer": "详细回答内容",
  "follow_ups": ["后续问题1", "后续问题2"]
}`,
			contextText, query)},
	}

	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM 生成回答失败: %w", err)
	}

	content := cleanJSON(resp.Content)
	log.Printf("[知识RAG] Step3: LLM 返回长度=%d", len(content))

	var result struct {
		Answer   string   `json:"answer"`
		FollowUp []string `json:"follow_ups"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// JSON 解析失败，返回原始内容
		log.Printf("[知识RAG] JSON 解析失败, 返回原始内容")
		return &KnowledgeAnswer{
			Answer:   content,
			Sources:  sources,
			FollowUp: []string{},
		}, nil
	}

	log.Printf("[知识RAG] 回答完成, follow_ups=%d", len(result.FollowUp))
	return &KnowledgeAnswer{
		Answer:    result.Answer,
		Sources:   sources,
		FollowUp:  result.FollowUp,
	}, nil
}
