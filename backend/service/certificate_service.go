package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"codelearn/model"
	"codelearn/repository"
)

var ErrAlreadyCertified = errors.New("该课程已有认证证书")

const (
	// CertLevelBeginnerThreshold 入门级证书最低分数
	CertLevelBeginnerThreshold = 60
	// CertLevelAdvancedThreshold 熟练级证书最低分数
	CertLevelAdvancedThreshold = 80
)

// LevelForScore 根据认证考试分数映射能力等级
func LevelForScore(score int) string {
	switch {
	case score >= CertLevelAdvancedThreshold:
		return "advanced"
	case score >= CertLevelBeginnerThreshold:
		return "intermediate"
	default:
		return "beginner"
	}
}

type CertificateService struct {
	repo *repository.Repository
}

func NewCertificateService(repo *repository.Repository) *CertificateService {
	return &CertificateService{repo: repo}
}

// IssueCertificate 颁发证书（同一用户同一课程仅可持有一张）
func (s *CertificateService) IssueCertificate(userID, courseID uint, score int) (*model.Certificate, error) {
	cnt, err := s.repo.CountCertificates(userID, courseID)
	if err != nil {
		return nil, err
	}
	if cnt > 0 {
		return nil, ErrAlreadyCertified
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	course, err := s.repo.GetCourse(courseID)
	if err != nil {
		return nil, err
	}

	cert := &model.Certificate{
		UserID:      userID,
		CourseID:    courseID,
		CertNo:      generateCertNo(),
		Level:       LevelForScore(score),
		Score:       score,
		CourseTitle: course.Title,
		UserName:    user.Username,
		IssuedAt:    time.Now(),
	}
	if err := s.repo.CreateCertificate(cert); err != nil {
		return nil, err
	}
	return cert, nil
}

// Verify 证书在线验证（公开接口，无需登录）
func (s *CertificateService) Verify(certNo string) (*model.Certificate, error) {
	return s.repo.GetCertificateByNo(strings.TrimSpace(certNo))
}

// ListByUser 我的证书列表
func (s *CertificateService) ListByUser(userID uint) ([]model.Certificate, error) {
	return s.repo.ListCertificatesByUser(userID)
}

// generateCertNo 生成不可枚举的证书编号：CL-YYYYMMDD-8位随机大写十六进制
func generateCertNo() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// 极端情况回退：用纳秒时间戳保证唯一
		return fmt.Sprintf("CL-%s-%d", time.Now().UTC().Format("20060102"), time.Now().UnixNano())
	}
	randHex := strings.ToUpper(hex.EncodeToString(b))
	return fmt.Sprintf("CL-%s-%s", time.Now().UTC().Format("20060102"), randHex)
}
