package auth

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/anomalyco/story/internal/infrastructure/config"
)

type JWTTokenService struct {
	secret         []byte
	accessTokenTTL time.Duration
}

func NewJWTTokenService(cfg config.AuthConfig) (*JWTTokenService, error) {
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT secret must be at least 32 characters")
	}
	return &JWTTokenService{
		secret:         []byte(cfg.JWTSecret),
		accessTokenTTL: cfg.AccessTokenTTL,
	}, nil
}

type CustomClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	SessionID uuid.UUID `json:"session_id"`
	jwt.RegisteredClaims
}

func (s *JWTTokenService) GenerateAccessToken(userID, sessionID uuid.UUID) (string, int64, error) {
	now := time.Now()
	claims := CustomClaims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "story",
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return "", 0, fmt.Errorf("signing token: %w", err)
	}

	return signedToken, int64(s.accessTokenTTL.Seconds()), nil
}

func (s *JWTTokenService) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating refresh token: %w", err)
	}
	return fmt.Sprintf("%x", b), nil
}

func (s *JWTTokenService) ValidateAccessToken(tokenString string) (uuid.UUID, uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid token claims")
	}

	return claims.UserID, claims.SessionID, nil
}
