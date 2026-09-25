package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"codelearn/config"

	"github.com/cloudwego/eino/schema"
	openaillm "github.com/cloudwego/eino-ext/components/model/openai"
)

// PlanAdvisor 使用 Eino 生成个性化学习计划（U5）：用户目标 → 周计划拆解
type PlanAdvisor struct {
	cfg *config.Config
}

func NewPlanAdvisor(cfg *config.Config) *PlanAdvisor {
	return &PlanAdvisor{cfg: cfg}
}

// Enabled 是否配置了 LLM（未配置则走模板降级）
func (a *PlanAdvisor) Enabled() bool {
	return a.cfg != nil && a.cfg.LLMAPIKey != ""
}

// PlanWeek 单周计划
type PlanWeek struct {
	Week  int      `json:"week"`
	Focus string   `json:"focus"`
	Tasks []string `json:"tasks"`
}

// StudyPlan LLM 生成的学习计划
type StudyPlan struct {
	Goal       string     `json:"goal"`
	Language   string     `json:"language"`
	TotalWeeks int        `json:"total_weeks"`
	Weeks      []PlanWeek `json:"weeks"`
	Summary    string     `json:"summary"`
}

// Generate 调用 LLM 生成学习计划
func (a *PlanAdvisor) Generate(ctx context.Context, goal, language string, weeks int) (*StudyPlan, error) {
	chatModel, err := openaillm.NewChatModel(ctx, &openaillm.ChatModelConfig{
		APIKey:  a.cfg.LLMAPIKey,
		BaseURL: a.cfg.LLMBaseURL,
		Model:   a.cfg.LLMModel,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	messages := []*schema.Message{
		{Role: schema.System, Content: `你是一名编程学习规划专家。根据用户的学习目标与周期，生成可执行的周学习计划。
要求：
1. 每周一个主题 focus，聚焦一个知识点阶段
2. 每周 5 条具体任务（学习+练习+复习结合，语言为中文）
3. 目标拆解循序渐进，适合零基础或初级学员
4. 严格返回 JSON 格式`},
		{Role: schema.User, Content: fmt.Sprintf(`用户目标：%s
编程语言：%s
计划周期：%d 周

请生成学习计划，返回 JSON：
{
  "goal": "目标",
  "language": "%s",
  "total_weeks": %d,
  "weeks": [
    {"week": 1, "focus": "本周主题", "tasks": ["任务1", "任务2", "任务3", "任务4", "任务5"]}
  ],
  "summary": "整体规划说明"
}`, goal, language, weeks, language, weeks)},
	}

	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM 生成计划失败: %w", err)
	}

	content := cleanJSON(resp.Content)
	log.Printf("[学习计划] LLM 返回长度=%d", len(content))

	var plan StudyPlan
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return nil, fmt.Errorf("解析计划失败: %w (原始: %s)", err, truncate(content, 200))
	}
	if len(plan.Weeks) == 0 {
		return nil, fmt.Errorf("计划为空")
	}
	return &plan, nil
}
