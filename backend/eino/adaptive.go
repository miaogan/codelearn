package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"codelearn/config"
	domainmodel "codelearn/model"

	"github.com/cloudwego/eino/schema"
	openaillm "github.com/cloudwego/eino-ext/components/model/openai"
)

// AdaptiveAdvisor 使用 Eino 的 Chain 模式编排多步骤 LLM 调用：
// Step1: 分析错题数据 → LLM 识别薄弱知识点
// Step2: 根据薄弱点 → LLM 生成针对性练习题
// 这展示了 Eino Chain 的核心思想：多个组件串联，数据在步骤间流转
type AdaptiveAdvisor struct {
	cfg *config.Config
}

func NewAdaptiveAdvisor(cfg *config.Config) *AdaptiveAdvisor {
	return &AdaptiveAdvisor{cfg: cfg}
}

type WeakPoint struct {
	Topic       string `json:"topic"`
	Description string `json:"description"`
	Count       int    `json:"count"`
}

type AdaptiveRecommendation struct {
	WeakPoints    []WeakPoint    `json:"weak_points"`
	Summary       string         `json:"summary"`
	Exercises     []domainmodel.Exercise `json:"exercises"`
}

func (a *AdaptiveAdvisor) createChatModel(ctx context.Context) (*openaillm.ChatModel, error) {
	return openaillm.NewChatModel(ctx, &openaillm.ChatModelConfig{
		APIKey:  a.cfg.LLMAPIKey,
		BaseURL: a.cfg.LLMBaseURL,
		Model:   a.cfg.LLMModel,
	})
}

// Analyze 使用 Eino Chain 模式：Step1 分析薄弱知识点
func (a *AdaptiveAdvisor) Analyze(ctx context.Context, wrongExercises []domainmodel.Exercise, language string) ([]WeakPoint, string, error) {
	log.Printf("[自适应] Step1: 分析错题, 共 %d 道", len(wrongExercises))

	// 构建错题摘要
	wrongSummary := ""
	for i, ex := range wrongExercises {
		wrongSummary += fmt.Sprintf("%d. [%s] %s\n", i+1, ex.Type, truncate(ex.Question, 80))
	}

	chatModel, err := a.createChatModel(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	messages := []*schema.Message{
		{Role: schema.System, Content: `你是一名编程教育分析专家。根据学生的错题记录，分析其薄弱知识点。
要求：
1. 识别 2-4 个薄弱知识点（按错误频率排序）
2. 给出每个知识点的简要说明
3. 提供一句话的总体学习建议
4. 严格返回 JSON 格式`},
		{Role: schema.User, Content: fmt.Sprintf(`学生正在学习 %s 语言，以下是错题记录：

%s

请分析薄弱知识点，返回 JSON：
{
  "weak_points": [
    {"topic": "知识点名称", "description": "为什么薄弱", "count": 出错次数}
  ],
  "summary": "总体学习建议"
}`, language, wrongSummary)},
	}

	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, "", fmt.Errorf("LLM 分析失败: %w", err)
	}

	content := cleanJSON(resp.Content)
	log.Printf("[自适应] Step1 完成, LLM 返回长度=%d", len(content))

	var result struct {
		WeakPoints []WeakPoint `json:"weak_points"`
		Summary    string      `json:"summary"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, "", fmt.Errorf("解析分析结果失败: %w (原始: %s)", err, truncate(content, 200))
	}

	log.Printf("[自适应] Step1: 识别 %d 个薄弱知识点", len(result.WeakPoints))
	return result.WeakPoints, result.Summary, nil
}

// GenerateExercises Eino Chain Step2: 根据薄弱知识点生成针对性练习
func (a *AdaptiveAdvisor) GenerateExercises(ctx context.Context, weakPoints []WeakPoint, language string) ([]domainmodel.Exercise, error) {
	log.Printf("[自适应] Step2: 针对薄弱点生成练习")

	weakTopics := ""
	for i, wp := range weakPoints {
		weakTopics += fmt.Sprintf("%d. %s: %s\n", i+1, wp.Topic, wp.Description)
	}

	chatModel, err := a.createChatModel(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	messages := []*schema.Message{
		{Role: schema.System, Content: fmt.Sprintf(`你是一名编程教育专家。根据学生的薄弱知识点，为 %s 语言生成针对性的练习题。
每个薄弱知识点生成 1-2 道选择题，共 3-5 道题。
题目要直接针对薄弱点进行强化训练。
严格返回 JSON 数组格式`+exerciseFormatSpec(), language)},
		{Role: schema.User, Content: fmt.Sprintf(`薄弱知识点：
%s

请生成针对性练习题。`, weakTopics)},
	}

	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM 生成练习失败: %w", err)
	}

	content := cleanJSON(resp.Content)
	log.Printf("[自适应] Step2 完成, LLM 返回长度=%d", len(content))

	var genExercises []GenExercise
	if err := json.Unmarshal([]byte(content), &genExercises); err != nil {
		return nil, fmt.Errorf("解析练习题失败: %w", err)
	}

	exercises := make([]domainmodel.Exercise, 0, len(genExercises))
	for i, ge := range genExercises {
		ex := domainmodel.Exercise{
			Type:        ge.Type,
			Question:    ge.Question,
			Answer:      ge.Answer,
			Explanation: ge.Explanation,
			Difficulty:  pickDifficulty(ge.Difficulty, "easy"),
			IsAIGen:     true,
			Order:       i,
		}
		if len(ge.Options) > 0 {
			opts, _ := json.Marshal(ge.Options)
			ex.Options = string(opts)
		}
		exercises = append(exercises, ex)
	}

	log.Printf("[自适应] Step2: 生成 %d 道练习题", len(exercises))
	return exercises, nil
}

// Recommend 完整的 Eino Chain：Analyze → GenerateExercises
func (a *AdaptiveAdvisor) Recommend(ctx context.Context, wrongExercises []domainmodel.Exercise, language string) (*AdaptiveRecommendation, error) {
	if a.cfg.LLMAPIKey == "" {
		return nil, fmt.Errorf("LLM_API_KEY 未设置")
	}

	// Chain Step1: 分析薄弱知识点
	weakPoints, summary, err := a.Analyze(ctx, wrongExercises, language)
	if err != nil {
		return nil, fmt.Errorf("分析阶段失败: %w", err)
	}

	if len(weakPoints) == 0 {
		return &AdaptiveRecommendation{
			WeakPoints: []WeakPoint{},
			Summary:    "暂无足够错题数据进行分析，继续学习以积累数据",
			Exercises:  []domainmodel.Exercise{},
		}, nil
	}

	// Chain Step2: 生成针对性练习
	exercises, err := a.GenerateExercises(ctx, weakPoints, language)
	if err != nil {
		return &AdaptiveRecommendation{
			WeakPoints: weakPoints,
			Summary:    summary,
			Exercises:  []domainmodel.Exercise{},
		}, nil
	}

	return &AdaptiveRecommendation{
		WeakPoints: weakPoints,
		Summary:    summary,
		Exercises:   exercises,
	}, nil
}

func exerciseFormatSpec() string {
	return `
返回格式：
[
  {
    "type": "choice",
    "question": "题目内容",
    "options": ["选项A", "选项B", "选项C", "选项D"],
    "answer": "正确选项的完整文本",
    "explanation": "解析说明",
    "difficulty": "easy"
  }
]`
}
