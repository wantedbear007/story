package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

// PasswordHasher abstracts password hashing for the auth service.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}

// TokenService abstracts JWT token operations for the auth service.
type TokenService interface {
	GenerateAccessToken(userID uuid.UUID) (string, error)
	GenerateRefreshToken() (string, error)
	ValidateAccessToken(tokenString string) (uuid.UUID, error)
}

// EmailSender abstracts sending emails (password reset, notifications).
type EmailSender interface {
	SendPasswordResetEmail(ctx context.Context, email, token string) error
}

// Service implements authentication and authorization use cases.
// It bridges user management, session management, and token operations
// without exposing infrastructure concerns to callers.
type Service struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
	hasher      PasswordHasher
	tokens      TokenService
	mailer      EmailSender
}

func NewService(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	hasher PasswordHasher,
	tokens TokenService,
	mailer EmailSender,
) *Service {
	return &Service{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		hasher:      hasher,
		tokens:      tokens,
		mailer:      mailer,
	}
}

func (s *Service) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*RefreshTokenResponse, error) {
	hash, err := s.hasher.Hash(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("processing refresh token: %w", err)
	}

	storedToken, err := s.sessionRepo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid refresh token", domain.ErrUnauthorized)
	}

	if !storedToken.IsValid() {
		return nil, fmt.Errorf("%w: refresh token expired or revoked", domain.ErrUnauthorized)
	}

	_ = s.sessionRepo.RevokeRefreshToken(ctx, storedToken.ID)

	user, err := s.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: user not found", domain.ErrNotFound)
	}

	accessToken, err := s.tokens.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	rawRefresh, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	newHash, err := s.hasher.Hash(rawRefresh)
	if err != nil {
		return nil, fmt.Errorf("hashing refresh token: %w", err)
	}

	newRefreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: newHash,
	}

	if err := s.sessionRepo.CreateRefreshToken(ctx, newRefreshToken); err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	return &RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
	}, nil
}

func (s *Service) RevokeToken(ctx context.Context, userID uuid.UUID, req RevokeTokenRequest) error {
	hash, err := s.hasher.Hash(req.RefreshToken)
	if err != nil {
		return fmt.Errorf("processing refresh token: %w", err)
	}

	storedToken, err := s.sessionRepo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return nil
	}

	if storedToken.UserID != userID {
		return nil
	}

	return s.sessionRepo.RevokeRefreshToken(ctx, storedToken.ID)
}

func (s *Service) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.RevokeUserRefreshTokens(ctx, userID)
}

func (s *Service) RequestPasswordReset(ctx context.Context, req RequestPasswordResetRequest) error {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil
	}

	rawToken := uuid.New().String()
	tokenHash, err := s.hasher.Hash(rawToken)
	if err != nil {
		return fmt.Errorf("hashing reset token: %w", err)
	}

	resetToken := &domain.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := s.sessionRepo.CreatePasswordResetToken(ctx, resetToken); err != nil {
		return fmt.Errorf("storing reset token: %w", err)
	}

	if err := s.mailer.SendPasswordResetEmail(ctx, user.Email, rawToken); err != nil {
		return fmt.Errorf("sending reset email: %w", err)
	}

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	hash, err := s.hasher.Hash(req.Token)
	if err != nil {
		return fmt.Errorf("processing reset token: %w", err)
	}

	storedToken, err := s.sessionRepo.GetPasswordResetTokenByHash(ctx, hash)
	if err != nil {
		return fmt.Errorf("%w: invalid or expired reset token", domain.ErrInvalidInput)
	}

	if storedToken.UsedAt != nil {
		return fmt.Errorf("%w: reset token already used", domain.ErrInvalidInput)
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return fmt.Errorf("%w: reset token expired", domain.ErrInvalidInput)
	}

	user, err := s.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return fmt.Errorf("%w: user not found", domain.ErrNotFound)
	}

	newHash, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		return fmt.Errorf("hashing new password: %w", err)
	}

	user.PasswordHash = newHash
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	if err := s.sessionRepo.MarkPasswordResetTokenUsed(ctx, storedToken.ID); err != nil {
		return fmt.Errorf("marking token as used: %w", err)
	}

	if err := s.sessionRepo.RevokeUserRefreshTokens(ctx, user.ID); err != nil {
		return fmt.Errorf("revoking sessions: %w", err)
	}

	return nil
}
