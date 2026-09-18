package models

// User はユーザーテーブルのモデル
type User struct {
	ID           string  `json:"id"`
	CompanyID    *string `json:"company_id"`
	Email        string  `json:"email"`
	PasswordHash string  `json:"password_hash"` // DB から読む際は必要、レスポンスには含めない
	Role         string  `json:"role"`
	DisplayName  *string `json:"display_name"`
	IsActive     bool    `json:"is_active"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// LoginRequest はログインリクエスト
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=3"`
}

// RegisterRequest は新規ユーザー登録リクエスト（管理者用）
type RegisterRequest struct {
	Email       string  `json:"email" binding:"required,email"`
	Password    string  `json:"password" binding:"required,min=6"`
	DisplayName *string `json:"display_name"`
	CompanyName string  `json:"company_name" binding:"required"` // 新規会社名
	Role        string  `json:"role"` // "admin" or "user"
}

// LoginResponse はログインレスポンス
type LoginResponse struct {
	Token       string `json:"token"`
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	CompanyID   string `json:"company_id"`
	DisplayName string `json:"display_name"`
}

// JWTClaims は JWT のペイロード
type JWTClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CompanyID string `json:"company_id"`
}
