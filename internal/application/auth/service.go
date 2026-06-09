package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/anomalyco/story/internal/domain"
	apperrors "github.com/anomalyco/story/internal/pkg/errors"
)

type TokenService interface {
	GenerateAccessToken(userID, sessionID uuid.UUID) (string, int64, error)
	GenerateRefreshToken() (string, error)
	ValidateAccessToken(tokenString string) (userID, sessionID uuid.UUID, err error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, encodedHash string) error
	HashToken(token string) string
}

type EmailSender interface {
	SendPasswordResetEmail(ctx context.Context, to, token, displayName string) error
}

type Service struct {
	userRepo         domain.UserRepository
	sessionRepo      domain.SessionRepository
	passwordRepo     domain.PasswordResetRepository
	tokenSvc         TokenService
	hasher           PasswordHasher
	mailer           EmailSender
	refreshTokenTTL  time.Duration
	passwordResetTTL time.Duration
}

func NewService(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	passwordRepo domain.PasswordResetRepository,
	tokenSvc TokenService,
	hasher PasswordHasher,
	mailer EmailSender,
	refreshTokenTTL time.Duration,
	passwordResetTTL time.Duration,
) *Service {
	return &Service{
		userRepo:         userRepo,
		sessionRepo:      sessionRepo,
		passwordRepo:     passwordRepo,
		tokenSvc:         tokenSvc,
		hasher:           hasher,
		mailer:           mailer,
		refreshTokenTTL:  refreshTokenTTL,
		passwordResetTTL: passwordResetTTL,
	}
}

func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperrors.ErrUnauthorized("invalid email or password")
		}
		return nil, fmt.Errorf("fetching user: %w", err)
	}

	if err := s.hasher.Verify(req.Password, user.PasswordHash); err != nil {
		return nil, apperrors.ErrUnauthorized("invalid email or password")
	}

	rawToken, err := s.tokenSvc.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	session := &domain.Session{
		ID:         uuid.New(),
		UserID:     user.ID,
		TokenHash:  s.hasher.HashToken(rawToken),
		DeviceInfo: req.DeviceInfo,
		IPAddress:  req.IPAddress,
		ExpiresAt:  time.Now().Add(s.refreshTokenTTL),
		LastUsedAt: timePtr(time.Now()),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	accessToken, expiresIn, err := s.tokenSvc.GenerateAccessToken(user.ID, session.ID)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawToken,
		SessionID:    session.ID.String(),
		User: &UserInfo{
			ID:          user.ID.String(),
			Email:       user.Email,
			DisplayName: user.DisplayName,
		},
		ExpiresIn: expiresIn,
	}, nil
}

func (s *Service) Logout(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("fetching session: %w", err)
	}

	if session.IsRevoked {
		return nil
	}

	return s.sessionRepo.Revoke(ctx, sessionID)
}

func (s *Service) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	tokenHash := s.hasher.HashToken(req.RefreshToken)

	oldSession, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperrors.ErrUnauthorized("invalid refresh token")
		}
		return nil, fmt.Errorf("fetching session: %w", err)
	}

	if !oldSession.IsActive() {
		return nil, apperrors.ErrUnauthorized("session is expired or revoked")
	}

	rawToken, err := s.tokenSvc.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	newSession := &domain.Session{
		ID:         uuid.New(),
		UserID:     oldSession.UserID,
		TokenHash:  s.hasher.HashToken(rawToken),
		DeviceInfo: oldSession.DeviceInfo,
		IPAddress:  oldSession.IPAddress,
		ExpiresAt:  time.Now().Add(s.refreshTokenTTL),
		LastUsedAt: timePtr(time.Now()),
	}

	if err := s.sessionRepo.Create(ctx, newSession); err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	if err := s.sessionRepo.Revoke(ctx, oldSession.ID); err != nil {
		return nil, fmt.Errorf("revoking old session: %w", err)
	}

	accessToken, expiresIn, err := s.tokenSvc.GenerateAccessToken(oldSession.UserID, newSession.ID)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	return &RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: rawToken,
		SessionID:    newSession.ID.String(),
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *Service) ListSessions(ctx context.Context, userID, currentSessionID uuid.UUID) ([]*SessionResponse, error) {
	sessions, err := s.sessionRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}

	result := make([]*SessionResponse, 0, len(sessions))
	currentIDStr := currentSessionID.String()
	for _, sess := range sessions {
		if sess.IsRevoked {
			continue
		}
		result = append(result, &SessionResponse{
			ID:         sess.ID.String(),
			DeviceInfo: sess.DeviceInfo,
			IPAddress:  sess.IPAddress,
			IsCurrent:  sess.ID.String() == currentIDStr,
			CreatedAt:  sess.CreatedAt,
			LastUsedAt: sess.LastUsedAt,
			ExpiresAt:  sess.ExpiresAt,
		})
	}
	return result, nil
}

func (s *Service) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperrors.ErrNotFound("session not found")
		}
		return fmt.Errorf("fetching session: %w", err)
	}

	if session.UserID != userID {
		return apperrors.ErrForbidden("session does not belong to this user")
	}

	return s.sessionRepo.Revoke(ctx, sessionID)
}

func (s *Service) RevokeAllSessions(ctx context.Context, userID, excludeSessionID uuid.UUID) error {
	sessions, err := s.sessionRepo.ListByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("listing sessions: %w", err)
	}

	for _, sess := range sessions {
		if sess.ID == excludeSessionID {
			continue
		}
		if !sess.IsRevoked {
			if err := s.sessionRepo.Revoke(ctx, sess.ID); err != nil {
				return fmt.Errorf("revoking session %s: %w", sess.ID, err)
			}
		}
	}

	return nil
}

func (s *Service) RequestPasswordReset(ctx context.Context, req *ForgotPasswordRequest) error {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("generating reset token: %w", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)

	prt := &domain.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: s.hasher.HashToken(rawToken),
		ExpiresAt: time.Now().Add(s.passwordResetTTL),
	}

	if err := s.passwordRepo.Create(ctx, prt); err != nil {
		return fmt.Errorf("creating password reset: %w", err)
	}

	if err := s.mailer.SendPasswordResetEmail(ctx, user.Email, rawToken, user.DisplayName); err != nil {
		_ = err
	}

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	tokenHash := s.hasher.HashToken(req.Token)

	prt, err := s.passwordRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperrors.ErrInvalidInput("invalid or expired reset token")
		}
		return fmt.Errorf("fetching reset token: %w", err)
	}

	if !prt.IsValid() {
		return apperrors.ErrInvalidInput("reset token has expired or already used")
	}

	newHash, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		return fmt.Errorf("hashing new password: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, prt.UserID)
	if err != nil {
		return fmt.Errorf("fetching user: %w", err)
	}

	user.PasswordHash = newHash
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	if err := s.passwordRepo.MarkUsed(ctx, prt.ID); err != nil {
		return fmt.Errorf("marking token used: %w", err)
	}

	if err := s.sessionRepo.RevokeAllForUser(ctx, user.ID); err != nil {
		return fmt.Errorf("revoking sessions: %w", err)
	}

	return nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}
