package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type Manager struct {
	secret    []byte
	expiresAt time.Duration
	refresh   time.Duration
}

func NewManager(secret string, expire, refresh time.Duration) *Manager {
	return &Manager{
		secret:    []byte(secret),
		expiresAt: expire,
		refresh:   refresh,
	}
}

func (m *Manager) GenerateAccessToken(userId, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"role":    role,
		"exp":     time.Now().Add(m.expiresAt).Unix(),
	})
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signed, nil
}

func (m *Manager) GenerateRefreshToken(userId string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(m.refresh).Unix(),
	})
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signed, nil
}

func (m *Manager) ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	role, _ := mapClaims["role"].(string)

	return &Claims{
		UserID: mapClaims["user_id"].(string),
		Role:   role,
	}, nil
}
