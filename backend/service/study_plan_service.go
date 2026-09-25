package service

import (
	"context"

	"codelearn/eino"
	"codelearn/repository"
)

// StudyPlanService 学习计划生成（U5）：LLM 生成 + 模板降级
type StudyPlanService struct {
	repo    *repository.Repository
	advisor *eino.PlanAdvisor
}

func NewStudyPlanService(repo *repository.Repository, advisor *eino.PlanAdvisor) *StudyPlanService {
	return &StudyPlanService{repo: repo, advisor: advisor}
}

// Generate 生成学习计划（LLM 优先，未配置或无 Key 时模板降级）
func (s *StudyPlanService) Generate(ctx context.Context, userID uint, goal, language string, weeks int) (*eino.StudyPlan, error) {
	if weeks < 1 {
		weeks = 4
	}
	if weeks > 12 {
		weeks = 12
	}
	if language == "" {
		language = "python"
	}
	if goal == "" {
		goal = "掌握 " + language + " 编程基础"
	}
	_ = userID

	if s.advisor != nil && s.advisor.Enabled() {
		if plan, err := s.advisor.Generate(ctx, goal, language, weeks); err == nil {
			return plan, nil
		}
	}
	return templatePlan(goal, language, weeks), nil
}

// templatePlan 模板降级：按周拆解基础知识点，无 LLM 依赖
func templatePlan(goal, language string, weeks int) *eino.StudyPlan {
	topics := []string{"基础语法", "变量与数据类型", "流程控制", "函数", "数据结构", "文件与错误处理", "综合练习", "项目实战", "复习巩固", "模拟考试", "冲刺提升", "认证备考"}
	weeksList := make([]eino.PlanWeek, 0, weeks)
	for w := 1; w <= weeks; w++ {
		topic := topics[(w-1)%len(topics)]
		weeksList = append(weeksList, eino.PlanWeek{
			Week:  w,
			Focus: topic,
			Tasks: []string{
				"学习《" + topic + "》相关课时（2 个课时）",
				"完成课时配套练习（3-5 题）",
				"整理错题并完成复习",
				"做 1 道编程实战小练习",
				"复习本周知识点并预习下周内容",
			},
		})
	}
	return &eino.StudyPlan{
		Goal:       goal,
		Language:   language,
		TotalWeeks: weeks,
		Weeks:      weeksList,
		Summary:    "围绕目标" + goal + "，按周拆解" + language + "核心知识点，学习-练习-复习闭环推进。",
	}
}
