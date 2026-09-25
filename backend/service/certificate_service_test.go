package service

import (
	"errors"
	"testing"
)

func TestIssueCertificate(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewCertificateService(repo)

	user := seedUser(db)
	courseID, _, _ := seedCourse(db)

	cert, err := svc.IssueCertificate(user.ID, courseID, 75)
	if err != nil {
		t.Fatalf("IssueCertificate failed: %v", err)
	}
	if cert.CertNo == "" {
		t.Error("expected cert no to be set")
	}
	if cert.Level != "intermediate" {
		t.Errorf("expected level intermediate (75), got %s", cert.Level)
	}
	if cert.UserName != "testuser" {
		t.Errorf("expected username snapshot, got %s", cert.UserName)
	}
}

func TestIssueCertificate_Duplicate(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewCertificateService(repo)

	user := seedUser(db)
	courseID, _, _ := seedCourse(db)

	if _, err := svc.IssueCertificate(user.ID, courseID, 70); err != nil {
		t.Fatalf("first issue failed: %v", err)
	}
	_, err := svc.IssueCertificate(user.ID, courseID, 90)
	if !errors.Is(err, ErrAlreadyCertified) {
		t.Fatalf("expected ErrAlreadyCertified, got %v", err)
	}
}

func TestVerifyCertificate(t *testing.T) {
	db := setupTestDB(t)
	repo := newTestRepo(db)
	svc := NewCertificateService(repo)

	user := seedUser(db)
	courseID, _, _ := seedCourse(db)

	cert, _ := svc.IssueCertificate(user.ID, courseID, 88)

	// 正确编号可验证
	verified, err := svc.Verify(cert.CertNo)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if verified.CertNo != cert.CertNo {
		t.Errorf("cert no mismatch")
	}

	// 伪造编号不可验证
	if _, err := svc.Verify("CL-20260101-DEADBEEF"); err == nil {
		t.Error("expected error for fake cert no")
	}
}

func TestLevelForScore(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{95, "advanced"},
		{80, "advanced"},
		{79, "intermediate"},
		{60, "intermediate"},
		{59, "beginner"},
	}
	for _, tt := range tests {
		if got := LevelForScore(tt.score); got != tt.want {
			t.Errorf("LevelForScore(%d) = %s, want %s", tt.score, got, tt.want)
		}
	}
}

func TestGenerateCertNo_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		no := generateCertNo()
		if seen[no] {
			t.Fatalf("duplicate cert no generated: %s", no)
		}
		seen[no] = true
	}
}
