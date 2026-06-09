package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/anomalyco/story/internal/infrastructure/config"
)

// JWTTokenService manages JWT access and refresh tokens.
// Access tokens are short-lived (15min default) and use HMAC-SHA256 signing.
// Refresh tokens are opaque random strings stored as Argon2 hashes.
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

// CustomClaims extends standard JWT claims with user identification.
type CustomClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a signed JWT for the given user.
func (s *JWTTokenService) GenerateAccessToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := CustomClaims{
		UserID: userID,
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
		return "", fmt.Errorf("signing token: %w", err)
	}

	return signedToken, nil
}

// GenerateRefreshToken creates a cryptographically random refresh token.
// The raw token should be returned to the client; only its hash is stored.
func (s *JWTTokenService) GenerateRefreshToken() (string, error) {
	b, err := generateRandomBytes(32)
	if err != nil {
		return "", fmt.Errorf("generating refresh token: %w", err)
	}
	return fmt.Sprintf("%x", b), nil
}

// ValidateAccessToken parses and validates a JWT, returning the user ID.
func (s *JWTTokenService) ValidateAccessToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token claims")
	}

	return claims.UserID, nil
}
