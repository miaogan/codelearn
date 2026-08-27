package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"codelearn/config"
	"codelearn/sandbox"

	"github.com/cloudwego/eino/schema"
	openaillm "github.com/cloudwego/eino-ext/components/model/openai"
)

// TutorAgent 使用 Eino 的 Agent 模式：LLM + 工具调用循环
//
// Agent 循环流程：
// 1. 学生提交有 bug 的代码 + 问题描述
// 2. Agent 调用 code_runner 工具运行代码
// 3. Agent 分析运行结果/错误
// 4. Agent 给出诊断 + 修复建议
// 5. 如果需要，Agent 可以再次运行修复后的代码验证
//
// 这展示了 Eino Agent 的核心思想：LLM 自主决定调用哪个工具，根据工具结果继续推理
type TutorAgent struct {
	cfg *config.Config
}

type TutorRequest struct {
	Language string `json:"language"`
	Code     string `json:"code"`
	Question string `json:"question"`
}

type TutorResponse struct {
	Diagnosis string `json:"diagnosis"`
	Suggest   string `json:"suggestion"`
	Fix       string `json:"fix_code"`
	RunResult string `json:"run_result"`
}

func NewTutorAgent(cfg *config.Config) *TutorAgent {
	return &TutorAgent{cfg: cfg}
}

// Debug 执行 Agent 循环：运行代码 → LLM 分析 → 给出诊断
func (t *TutorAgent) Debug(ctx context.Context, req TutorRequest) (*TutorResponse, error) {
	log.Printf("[AI导师] 收到请求: language=%s question=%s code_len=%d", req.Language, truncate(req.Question, 50), len(req.Code))

	if t.cfg.LLMAPIKey == "" {
		return nil, fmt.Errorf("LLM_API_KEY 未设置")
	}

	// Agent Tool 1: 运行学生代码
	log.Printf("[AI导师] Tool: 运行代码")
	runResult := sandbox.RunCode(req.Language, req.Code)
	runOutput := ""
	if runResult.Error != "" {
		runOutput = runResult.Error
		log.Printf("[AI导师] 代码运行出错: %s", truncate(runResult.Error, 100))
	} else {
		runOutput = runResult.Output
		log.Printf("[AI导师] 代码运行成功: %s", truncate(runResult.Output, 100))
	}

	// Agent LLM: 分析代码 + 运行结果，给出诊断
	log.Printf("[AI导师] LLM: 分析代码并诊断")
	chatModel, err := openaillm.NewChatModel(ctx, &openaillm.ChatModelConfig{
		APIKey:  t.cfg.LLMAPIKey,
		BaseURL: t.cfg.LLMBaseURL,
		Model:   t.cfg.LLMModel,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	messages := []*schema.Message{
		{Role: schema.System, Content: `你是一名编程导师，擅长诊断和修复代码 bug。

你的工作流程：
1. 分析学生提交的代码和运行结果
2. 找出 bug 的根本原因
3. 解释为什么会出现这个 bug
4. 给出修复后的完整代码
5. 如果有必要，解释修复的原理

返回严格的 JSON 格式，不要包含 markdown 标记`},
		{Role: schema.User, Content: fmt.Sprintf(`语言：%s
学生的问题：%s

学生的代码：
%s

代码运行结果：
%s

请诊断并修复，返回 JSON：
{
  "diagnosis": "bug 的根本原因分析",
  "suggestion": "学习建议，如何避免类似 bug",
  "fix_code": "修复后的完整代码",
  "run_result": "对运行结果的简要说明"
}`,
			req.Language, req.Question, req.Code, runOutput)},
	}

	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM 诊断失败: %w", err)
	}

	content := cleanJSON(resp.Content)
	log.Printf("[AI导师] LLM 返回长度=%d", len(content))

	var result TutorResponse
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// 如果 JSON 解析失败，直接返回 LLM 原始内容
		log.Printf("[AI导师] JSON 解析失败, 返回原始内容: %v", err)
		return &TutorResponse{
			Diagnosis: "解析失败，LLM 原始回复",
			Suggest:   "",
			Fix:       "",
			RunResult: content,
		}, nil
	}

	log.Printf("[AI导师] 诊断完成: diagnosis=%s", truncate(result.Diagnosis, 80))
	return &result, nil
}

// Chat 多轮对话模式：学生可以追问
func (t *TutorAgent) Chat(ctx context.Context, messages []TutorChatMessage, code, language string) (string, error) {
	log.Printf("[AI导师] 多轮对话: %d 条消息", len(messages))

	if t.cfg.LLMAPIKey == "" {
		return "", fmt.Errorf("LLM_API_KEY 未设置")
	}

	chatModel, err := openaillm.NewChatModel(ctx, &openaillm.ChatModelConfig{
		APIKey:  t.cfg.LLMAPIKey,
		BaseURL: t.cfg.LLMBaseURL,
		Model:   t.cfg.LLMModel,
	})
	if err != nil {
		return "", fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	// 把对话历史转换为 Eino schema.Message
	einoMessages := []*schema.Message{
		{Role: schema.System, Content: fmt.Sprintf(`你是一名编程导师，正在帮助学生调试 %s 代码。

学生的当前代码：
%s

请根据对话上下文回答学生的问题。如果学生问关于代码的问题，请结合代码内容回答。
回答要简洁明了，适合编程初学者理解。`, language, code)},
	}

	for _, m := range messages {
		role := schema.User
		if m.Role == "assistant" {
			role = schema.Assistant
		}
		einoMessages = append(einoMessages, &schema.Message{
			Role:    role,
			Content: m.Content,
		})
	}

	resp, err := chatModel.Generate(ctx, einoMessages)
	if err != nil {
		return "", fmt.Errorf("LLM 对话失败: %w", err)
	}

	log.Printf("[AI导师] 对话回复长度=%d", len(resp.Content))
	return resp.Content, nil
}

type TutorChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnalyzeCode 使用 LLM 做代码审查（Agent Tool 模式的扩展）
func (t *TutorAgent) AnalyzeCode(ctx context.Context, code, language string) (string, error) {
	log.Printf("[AI导师] 代码审查: language=%s code_len=%d", language, len(code))

	if t.cfg.LLMAPIKey == "" {
		return "", fmt.Errorf("LLM_API_KEY 未设置")
	}

	chatModel, err := openaillm.NewChatModel(ctx, &openaillm.ChatModelConfig{
		APIKey:  t.cfg.LLMAPIKey,
		BaseURL: t.cfg.LLMBaseURL,
		Model:   t.cfg.LLMModel,
	})
	if err != nil {
		return "", fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	messages := []*schema.Message{
		{Role: schema.System, Content: fmt.Sprintf(`你是一名 %s 代码审查专家。请审查以下代码，从这些方面给出建议：

1. 代码规范（命名、格式）
2. 潜在 bug
3. 性能问题
4. 安全问题
5. 改进建议

用中文回答，简洁明了，每条建议不超过两句话。`, language)},
		{Role: schema.User, Content: fmt.Sprintf("请审查以下代码：\n\n%s", code)},
	}

	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("代码审查失败: %w", err)
	}

	log.Printf("[AI导师] 代码审查完成")
	return resp.Content, nil
}

// RunCodeForTutor 供前端调用的代码运行工具
func (t *TutorAgent) RunCodeForTutor(language, code, input string) *sandbox.RunResult {
	log.Printf("[AI导师] 运行代码: language=%s", language)
	result := sandbox.RunCode(language, code)
	return result
}

// truncate helper - 避免与 exercise_gen.go 中的 truncate 冲突
func init() {
	_ = strings.TrimSpace("")
	_ = fmt.Sprintf
}
