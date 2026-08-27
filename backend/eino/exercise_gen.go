package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"codelearn/config"
	domainmodel "codelearn/model"

	"github.com/cloudwego/eino/schema"
	openaillm "github.com/cloudwego/eino-ext/components/model/openai"
)

// ExerciseGenerator 使用 Eino 框架的 ChatModel 组件来生成编程练习题。
// Eino 是字节跳动开源的 Go LLM 应用开发框架，核心组件包括 ChatModel、
// ChatTemplate、Retriever、Tool 等，可通过 Chain/Graph 编排组合。
// 这里使用 ChatModel 直接调用 LLM 生成结构化习题。
type ExerciseGenerator struct {
	cfg *config.Config
}

func NewExerciseGenerator(cfg *config.Config) *ExerciseGenerator {
	return &ExerciseGenerator{cfg: cfg}
}

type GenRequest struct {
	Language   string `json:"language"`
	Topic      string `json:"topic"`
	Count      int    `json:"count"`
	Type       string `json:"type"`
	Difficulty string `json:"difficulty"`
}

type GenExercise struct {
	Type         string        `json:"type"`
	Question     string        `json:"question"`
	Options      []string      `json:"options,omitempty"`
	Answer       string        `json:"answer"`
	Explanation  string        `json:"explanation"`
	CodeTemplate string        `json:"code_template,omitempty"`
	TestCases    []GenTestCase `json:"test_cases,omitempty"`
	Difficulty   string        `json:"difficulty"`
}

type GenTestCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

func (g *ExerciseGenerator) buildSystemPrompt(req GenRequest) string {
	return fmt.Sprintf(`你是一名编程教育专家，擅长为初学者设计高质量的学习练习题。
请为「%s」编程语言的「%s」主题生成 %d 道「%s」难度的练习题。

习题类型说明：
- choice: 选择题，必须包含 4 个选项（options 数组），answer 是正确选项的完整文本
- fillblank: 填空题，answer 是正确答案文本
- subjective: 主观题/简答题，answer 是参考答案要点（用于 AI 判定学生答案），explanation 包含评分标准
- code: 代码编写题，必须包含 code_template（带函数签名或占位符）和 test_cases（测试用例数组，每个含 input 和 expected）
- order: 代码排序题，options 是打乱顺序的代码行，answer 是正确顺序（用换行连接的正确代码）

要求：
1. 题目内容准确，符合 %s 语言规范
2. 解释要清晰易懂，适合初学者
3. 难度为 %s 级别
4. 严格返回 JSON 数组格式，不要包含 markdown 标记或额外文字

返回格式示例：
[
  {
    "type": "choice",
    "question": "在 Go 中，哪个关键字用于声明常量？",
    "options": ["var", "const", "let", "final"],
    "answer": "const",
    "explanation": "Go 使用 const 关键字声明常量。",
    "difficulty": "easy"
  },
  {
    "type": "subjective",
    "question": "请简述 Go 语言中 goroutine 和 channel 的关系。",
    "answer": "goroutine 是轻量级线程，channel 是 goroutine 间通信的管道。通过 channel 可以实现 goroutine 间的数据传递和同步，避免共享内存带来的并发问题。",
    "explanation": "评分标准：1.提到goroutine是轻量级线程 2.提到channel用于通信 3.提到避免共享内存问题。满足2点以上为正确。",
    "difficulty": "medium"
  }
]`, req.Language, req.Topic, req.Count, req.Difficulty, req.Language, req.Difficulty)
}

func (g *ExerciseGenerator) buildUserPrompt(req GenRequest) string {
	return fmt.Sprintf("请为 %s 语言的「%s」主题生成 %d 道 %s 难度的「%s」类型练习题。直接返回 JSON 数组。",
		req.Language, req.Topic, req.Count, req.Difficulty, req.Type)
}

// Generate 调用 Eino ChatModel 生成习题并解析为数据库模型。
// Eino 的 ChatModel 是核心组件之一，通过 Generate 方法接收消息列表并返回 LLM 响应。
func (g *ExerciseGenerator) Generate(ctx context.Context, req GenRequest) ([]domainmodel.Exercise, error) {
	log.Printf("[AI生成] 开始: language=%s topic=%s count=%d type=%s difficulty=%s",
		req.Language, req.Topic, req.Count, req.Type, req.Difficulty)
	log.Printf("[AI生成] LLM配置: baseURL=%s model=%s apiKeyLen=%d",
		g.cfg.LLMBaseURL, g.cfg.LLMModel, len(g.cfg.LLMAPIKey))

	if g.cfg.LLMAPIKey == "" {
		log.Printf("[AI生成] 错误: LLM_API_KEY 未设置")
		return nil, fmt.Errorf("LLM_API_KEY 未设置，请配置环境变量 LLM_API_KEY")
	}

	// 使用 EinoExt 的 OpenAI 兼容模型实现创建 ChatModel
	log.Printf("[AI生成] 创建 ChatModel...")
	chatModel, err := openaillm.NewChatModel(ctx, &openaillm.ChatModelConfig{
		APIKey:  g.cfg.LLMAPIKey,
		BaseURL: g.cfg.LLMBaseURL,
		Model:   g.cfg.LLMModel,
	})
	if err != nil {
		log.Printf("[AI生成] 创建 ChatModel 失败: %v", err)
		return nil, fmt.Errorf("创建 ChatModel 失败: %w", err)
	}
	log.Printf("[AI生成] ChatModel 创建成功")

	// 使用 Eino 的 schema.Message 构造对话消息
	messages := []*schema.Message{
		{Role: schema.System, Content: g.buildSystemPrompt(req)},
		{Role: schema.User, Content: g.buildUserPrompt(req)},
	}

	// 调用 ChatModel.Generate
	log.Printf("[AI生成] 调用 LLM Generate...")
	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Printf("[AI生成] LLM 调用失败: %v", err)
		return nil, fmt.Errorf("调用 LLM 生成失败: %w", err)
	}

	content := cleanJSON(resp.Content)
	log.Printf("[AI生成] LLM 返回内容长度=%d, 前200字符: %s", len(content), truncate(content, 200))

	var genExercises []GenExercise
	if err := json.Unmarshal([]byte(content), &genExercises); err != nil {
		log.Printf("[AI生成] JSON解析失败: %v", err)
		log.Printf("[AI生成] 原始内容: %s", truncate(content, 500))
		return nil, fmt.Errorf("解析 LLM 返回的 JSON 失败: %w (原始内容: %s)", err, truncate(content, 200))
	}

	log.Printf("[AI生成] JSON解析成功, 共 %d 道习题", len(genExercises))

	exercises := make([]domainmodel.Exercise, 0, len(genExercises))
	for i, ge := range genExercises {
		log.Printf("[AI生成] 习题#%d: type=%s question=%s", i, ge.Type, truncate(ge.Question, 50))
		ex := domainmodel.Exercise{
			Type:        ge.Type,
			Question:    ge.Question,
			Answer:      ge.Answer,
			Explanation: ge.Explanation,
			Difficulty:  pickDifficulty(ge.Difficulty, req.Difficulty),
			IsAIGen:     true,
			Order:       i,
		}
		if len(ge.Options) > 0 {
			opts, _ := json.Marshal(ge.Options)
			ex.Options = string(opts)
		}
		if ge.CodeTemplate != "" {
			ex.CodeTemplate = ge.CodeTemplate
		}
		if len(ge.TestCases) > 0 {
			tc, _ := json.Marshal(ge.TestCases)
			ex.TestCases = string(tc)
		}
		exercises = append(exercises, ex)
	}

	log.Printf("[AI生成] 完成, 生成 %d 道习题", len(exercises))
	return exercises, nil
}

// GenerateHint 调用 LLM 为学生答错的题目提供学习提示（引导思考，不直接给答案）
func (g *ExerciseGenerator) GenerateHint(ctx context.Context, question, userAnswer, language string) (string, error) {
	log.Printf("[AI提示] 开始: language=%s question=%s userAnswer=%s",
		language, truncate(question, 50), truncate(userAnswer, 50))

	if g.cfg.LLMAPIKey == "" {
		log.Printf("[AI提示] 错误: LLM_API_KEY 未设置")
		return "", fmt.Errorf("LLM_API_KEY 未设置")
	}

	log.Printf("[AI提示] 创建 ChatModel...")
	chatModel, err := openaillm.NewChatModel(ctx, &openaillm.ChatModelConfig{
		APIKey:  g.cfg.LLMAPIKey,
		BaseURL: g.cfg.LLMBaseURL,
		Model:   g.cfg.LLMModel,
	})
	if err != nil {
		log.Printf("[AI提示] 创建 ChatModel 失败: %v", err)
		return "", fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	messages := []*schema.Message{
		{
			Role: schema.System,
			Content: fmt.Sprintf(`你是一名耐心的编程老师。学生在学习 %s 时答错了题目。
请给一个有帮助的提示，引导学生思考正确答案，但不要直接告诉答案。提示不超过两句话。
题目：%s
学生的回答：%s`, language, question, userAnswer),
		},
		{Role: schema.User, Content: "请给我一个提示。"},
	}

	log.Printf("[AI提示] 调用 LLM Generate...")
	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Printf("[AI提示] LLM 调用失败: %v", err)
		return "", fmt.Errorf("生成提示失败: %w", err)
	}

	log.Printf("[AI提示] 成功, 提示长度=%d", len(resp.Content))
	return resp.Content, nil
}

// JudgeSubjective 使用 LLM 判定主观题答案是否正确
// 返回: correct(bool), feedback(string), error
func (g *ExerciseGenerator) JudgeSubjective(ctx context.Context, question, referenceAnswer, studentAnswer, criteria string) (bool, string, error) {
	log.Printf("[AI判定] 开始: question=%s studentAnswer=%s", truncate(question, 50), truncate(studentAnswer, 50))

	if g.cfg.LLMAPIKey == "" {
		log.Printf("[AI判定] 错误: LLM_API_KEY 未设置")
		return false, "LLM_API_KEY 未设置", fmt.Errorf("LLM_API_KEY 未设置")
	}

	log.Printf("[AI判定] 创建 ChatModel...")
	chatModel, err := openaillm.NewChatModel(ctx, &openaillm.ChatModelConfig{
		APIKey:  g.cfg.LLMAPIKey,
		BaseURL: g.cfg.LLMBaseURL,
		Model:   g.cfg.LLMModel,
	})
	if err != nil {
		log.Printf("[AI判定] 创建 ChatModel 失败: %v", err)
		return false, "", fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	messages := []*schema.Message{
		{Role: schema.System, Content: `你是一名编程教育阅卷老师。你需要判定学生的主观题答案是否正确。

判定规则：
1. 不要求完全一致，只要核心概念正确即可
2. 学生答案只要包含参考答案中的关键要点，即为正确
3. 返回严格的 JSON 格式，不要包含 markdown 标记`},
		{Role: schema.User, Content: fmt.Sprintf(`题目：%s
参考答案：%s
评分标准：%s
学生的答案：%s

请判定学生答案是否正确，返回 JSON：
{"correct": true/false, "feedback": "简要说明判定理由，不超过两句话"}`,
			question, referenceAnswer, criteria, studentAnswer)},
	}

	log.Printf("[AI判定] 调用 LLM Generate...")
	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Printf("[AI判定] LLM 调用失败: %v", err)
		return false, "", fmt.Errorf("AI 判定失败: %w", err)
	}

	content := cleanJSON(resp.Content)
	log.Printf("[AI判定] LLM 返回: %s", truncate(content, 200))

	var result struct {
		Correct  bool   `json:"correct"`
		Feedback string `json:"feedback"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		log.Printf("[AI判定] JSON解析失败: %v, 原始: %s", err, truncate(content, 200))
		return false, "", fmt.Errorf("解析 AI 判定结果失败: %w", err)
	}

	log.Printf("[AI判定] 完成: correct=%v feedback=%s", result.Correct, truncate(result.Feedback, 50))
	return result.Correct, result.Feedback, nil
}

func cleanJSON(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func pickDifficulty(d, fallback string) string {
	switch strings.ToLower(d) {
	case "easy", "medium", "hard":
		return strings.ToLower(d)
	}
	return fallback
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
