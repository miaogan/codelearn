package service

import (
	"testing"
)

func TestCourseSkillMap_Statuses(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewSkillMapService(repo)

	user := seedUser(db)
	unitID := seedExamUnit(db, user.ID, true) // 已创建 course + unit + 2 lessons + 3 exercises
	u, _ := repo.GetUnit(unitID)

	skill, err := svc.CourseSkillMap(u.CourseID, user.ID)
	if err != nil {
		t.Fatalf("CourseSkillMap failed: %v", err)
	}
	if len(skill.Nodes) != 2 {
		t.Fatalf("expected 2 skill nodes, got %d", len(skill.Nodes))
	}

	// 课时已完成但无练习数据 => 学习中（非薄弱、非未学）
	for _, n := range skill.Nodes {
		if !n.Completed {
			t.Error("expected both lessons completed")
		}
		if n.Status != SkillStatusLearning {
			t.Errorf("expected 学习中 for completed lesson, got %s", n.Status)
		}
	}
	if skill.Overall < 40 || skill.Overall > 60 {
		t.Errorf("expected overall ~50, got %d", skill.Overall)
	}
}

func TestCourseSkillMap_WeakWithWrongs(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewSkillMapService(repo)

	user := seedUser(db)
	unitID := seedExamUnit(db, user.ID, true)
	u, _ := repo.GetUnit(unitID)

	// 给 lesson1 的一道题制造未掌握错题（错题惩罚 25）
	var lesson1ID uint
	lessons, _ := repo.ListLessonsByUnit(unitID)
	if len(lessons) > 0 {
		lesson1ID = lessons[0].ID
	}
	exs, _ := repo.ListExercisesByLesson(lesson1ID)
	if len(exs) == 0 {
		t.Fatal("no exercises for lesson1")
	}
	repo.UpsertWrongExercise(user.ID, exs[0].ID, "wrong", "exam")

	skill, err := svc.CourseSkillMap(u.CourseID, user.ID)
	if err != nil {
		t.Fatalf("CourseSkillMap failed: %v", err)
	}

	// lesson1 有错题惩罚 => 25 分 => 薄弱
	for _, n := range skill.Nodes {
		if n.LessonID == lesson1ID {
			if n.Status != SkillStatusWeak {
				t.Errorf("expected 薄弱 for lesson with wrongs, got %s (mastery=%d)", n.Status, n.Mastery)
			}
		}
	}
}

func TestReviewRecommendations(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewSkillMapService(repo)

	user := seedUser(db)
	unitID := seedExamUnit(db, user.ID, false)
	lessons, _ := repo.ListLessonsByUnit(unitID)
	exs, _ := repo.ListExercisesByLesson(lessons[0].ID)

	// 制造 lesson1 的 2 道错题
	if len(exs) < 2 {
		t.Fatal("need at least 2 exercises")
	}
	repo.UpsertWrongExercise(user.ID, exs[0].ID, "wrong1", "exercise")
	repo.UpsertWrongExercise(user.ID, exs[1].ID, "wrong2", "exercise")

	recs, err := svc.ReviewRecommendations(user.ID, 5)
	if err != nil {
		t.Fatalf("ReviewRecommendations failed: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(recs))
	}
	if recs[0].LessonID != lessons[0].ID {
		t.Errorf("expected lesson %d, got %d", lessons[0].ID, recs[0].LessonID)
	}
	if recs[0].WeakCount != 2 {
		t.Errorf("expected weak count 2, got %d", recs[0].WeakCount)
	}
	if recs[0].CourseTitle == "" {
		t.Error("expected course title populated")
	}
}

func TestReviewRecommendations_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewSkillMapService(repo)

	user := seedUser(db)
	seedExamUnit(db, user.ID, false)

	recs, err := svc.ReviewRecommendations(user.ID, 5)
	if err != nil {
		t.Fatalf("ReviewRecommendations failed: %v", err)
	}
	if len(recs) != 0 {
		t.Errorf("expected 0 recommendations, got %d", len(recs))
	}
}
