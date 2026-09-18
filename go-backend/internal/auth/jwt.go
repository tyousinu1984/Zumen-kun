package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zumen-kun/go-backend/internal/db/models"
)

var jwtSecret = []byte(getJWTSecret())

func getJWTSecret() string {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "zumen-kun-dev-secret-change-in-production"
	}
	return s
}

type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CompanyID string `json:"company_id"`
	jwt.RegisteredClaims
}

// GenerateToken はJWTトークンを生成する
func GenerateToken(user *models.User) (string, error) {
	companyID := ""
	if user.CompanyID != nil {
		companyID = *user.CompanyID
	}
	displayName := ""
	if user.DisplayName != nil {
		displayName = *user.DisplayName
	}
	_ = displayName

	claims := &Claims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		CompanyID: companyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken はJWTトークンを検証しクレームを返す
func ValidateToken(tokenStr string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &models.JWTClaims{
		UserID:    claims.UserID,
		Email:     claims.Email,
		Role:      claims.Role,
		CompanyID: claims.CompanyID,
	}, nil
}
