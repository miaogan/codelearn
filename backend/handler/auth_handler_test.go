package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"codelearn/middleware"
	"codelearn/model"
	"codelearn/repository"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB, *AuthHandler, *repository.Repository) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	db.AutoMigrate(&model.User{}, &model.Course{}, &model.Unit{}, &model.Lesson{}, &model.Exercise{}, &model.UserProgress{}, &model.Submission{})

	repo := repository.New(db)
	authHandler := NewAuthHandler(repo, "test-secret", 5)

	r := gin.New()
	api := r.Group("/api")
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	authed := api.Group("")
	authed.Use(middleware.AuthMiddleware("test-secret"))
	authed.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	return r, db, authHandler, repo
}

// registerUser 通过 HTTP 请求注册用户，返回 token
func registerUserViaHTTP(r *gin.Engine, username, email, password string) (string, int) {
	body, _ := json.Marshal(registerReq{
		Username: username, Email: email, Password: password,
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if token, ok := resp["token"].(string); ok {
		return token, w.Code
	}
	return "", w.Code
}

func TestRegister_Success(t *testing.T) {
	r, db, _, _ := setupTestRouter(t)

	token, code := registerUserViaHTTP(r, "newuser", "new@example.com", "password123")
	if code != http.StatusCreated {
		t.Errorf("expected 201, got %d", code)
	}
	if token == "" {
		t.Error("expected token in response")
	}

	var dbUser model.User
	db.First(&dbUser, "username = ?", "newuser")
	if dbUser.Hearts != 5 {
		t.Errorf("expected 5 hearts, got %d", dbUser.Hearts)
	}
	if dbUser.PasswordHash == "password123" {
		t.Error("password should be hashed")
	}
	if dbUser.MaxHearts != 5 {
		t.Errorf("expected maxHearts=5, got %d", dbUser.MaxHearts)
	}
}

func TestRegister_DuplicateUser(t *testing.T) {
	r, db, _, _ := setupTestRouter(t)

	user := model.User{
		Username: "existing", Email: "existing@example.com",
		PasswordHash: "hashed", Hearts: 5, MaxHearts: 5,
	}
	db.Create(&user)

	body, _ := json.Marshal(registerReq{
		Username: "existing", Email: "another@example.com", Password: "password123",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestRegister_InvalidInput(t *testing.T) {
	r, _, _, _ := setupTestRouter(t)

	tests := []struct {
		name string
		body string
	}{
		{"short password", `{"username":"ab","email":"a@b.com","password":"12"}`},
		{"missing email", `{"username":"test","password":"123456"}`},
		{"invalid email", `{"username":"test","email":"notemail","password":"123456"}`},
		{"short username", `{"username":"a","email":"a@b.com","password":"123456"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("%s: expected 400, got %d", tt.name, w.Code)
			}
		})
	}
}

func TestLogin_Success(t *testing.T) {
	r, _, _, _ := setupTestRouter(t)
	registerUserViaHTTP(r, "loginuser", "login@example.com", "password123")

	body, _ := json.Marshal(loginReq{Account: "loginuser", Password: "password123"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["token"] == nil {
		t.Error("expected token in response")
	}
}

func TestLogin_ByEmail(t *testing.T) {
	r, _, _, _ := setupTestRouter(t)
	registerUserViaHTTP(r, "emailuser", "email@example.com", "password123")

	body, _ := json.Marshal(loginReq{Account: "email@example.com", Password: "password123"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	r, _, _, _ := setupTestRouter(t)
	registerUserViaHTTP(r, "wrongpw", "wrong@example.com", "password123")

	body, _ := json.Marshal(loginReq{Account: "wrongpw", Password: "wrongpassword"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	r, _, _, _ := setupTestRouter(t)

	body, _ := json.Marshal(loginReq{Account: "nonexistent", Password: "password123"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	r, _, _, _ := setupTestRouter(t)
	token, _ := registerUserViaHTTP(r, "mwuser", "mw@example.com", "password123")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	r, _, _, _ := setupTestRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	r, _, _, _ := setupTestRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_WrongFormat(t *testing.T) {
	r, _, _, _ := setupTestRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Basic sometoken")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_EmptyAuth(t *testing.T) {
	r, _, _, _ := setupTestRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
