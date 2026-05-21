package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zumen-kun/go-backend/internal/auth"
	"github.com/zumen-kun/go-backend/internal/db"
	"github.com/zumen-kun/go-backend/internal/db/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db *db.SupabaseClient
}

func NewAuthHandler(db *db.SupabaseClient) *AuthHandler {
	return &AuthHandler{db: db}
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.db.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "メールアドレスまたはパスワードが正しくありません"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "このアカウントは無効化されています"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "メールアドレスまたはパスワードが正しくありません"})
		return
	}

	token, err := auth.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "トークン生成に失敗しました"})
		return
	}

	companyID := ""
	if user.CompanyID != nil {
		companyID = *user.CompanyID
	}
	displayName := ""
	if user.DisplayName != nil {
		displayName = *user.DisplayName
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:       token,
		UserID:      user.ID,
		Email:       user.Email,
		Role:        user.Role,
		CompanyID:   companyID,
		DisplayName: displayName,
	})
}

// POST /api/v1/auth/register  （管理者が新会社 + 最初のユーザーを作成）
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. 会社を作成
	companyID := uuid.New().String()
	if err := h.db.CreateCompany(c.Request.Context(), companyID, req.CompanyName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "会社の作成に失敗しました: " + err.Error()})
		return
	}

	// 2. パスワードをハッシュ化
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "パスワードのハッシュ化に失敗しました"})
		return
	}

	// 3. ユーザーを作成
	role := req.Role
	if role == "" {
		role = "admin"
	}
	user := &models.User{
		ID:           uuid.New().String(),
		CompanyID:    &companyID,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         role,
		DisplayName:  req.DisplayName,
		IsActive:     true,
	}

	if err := h.db.CreateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ユーザーの作成に失敗しました: " + err.Error()})
		return
	}

	token, err := auth.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "トークン生成に失敗しました"})
		return
	}

	displayName := ""
	if user.DisplayName != nil {
		displayName = *user.DisplayName
	}

	c.JSON(http.StatusCreated, models.LoginResponse{
		Token:       token,
		UserID:      user.ID,
		Email:       user.Email,
		Role:        user.Role,
		CompanyID:   companyID,
		DisplayName: displayName,
	})
}

// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	email, _ := c.Get("email")
	role, _ := c.Get("role")
	companyID, _ := c.Get("company_id")

	c.JSON(http.StatusOK, gin.H{
		"user_id":    userID,
		"email":      email,
		"role":       role,
		"company_id": companyID,
	})
}
