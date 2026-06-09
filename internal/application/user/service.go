package user

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

type EmailSender interface {
	SendVerificationEmail(ctx context.Context, to, token, displayName string) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, encodedHash string) error
	HashToken(token string) string
}

type Service struct {
	userRepo   domain.UserRepository
	emailRepo  domain.EmailVerificationRepository
	sessionRepo domain.SessionRepository
	hasher     PasswordHasher
	mailer     EmailSender
	verifyTTL  time.Duration
}

func NewService(
	userRepo domain.UserRepository,
	emailRepo domain.EmailVerificationRepository,
	sessionRepo domain.SessionRepository,
	hasher PasswordHasher,
	mailer EmailSender,
	verifyTTL time.Duration,
) *Service {
	return &Service{
		userRepo:    userRepo,
		emailRepo:   emailRepo,
		sessionRepo: sessionRepo,
		hasher:      hasher,
		mailer:      mailer,
		verifyTTL:   verifyTTL,
	}
}

func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	existing, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, apperrors.ErrConflict("email already registered")
	}

	pwHash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: pwHash,
		DisplayName:  req.DisplayName,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("generating verification token: %w", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)

	ev := &domain.EmailVerification{
		ID:        uuid.New(),
		UserID:    user.ID,
		Email:     user.Email,
		TokenHash: s.hasher.HashToken(rawToken),
		ExpiresAt: time.Now().Add(s.verifyTTL),
	}

	if err := s.emailRepo.Create(ctx, ev); err != nil {
		return nil, fmt.Errorf("creating email verification: %w", err)
	}

	if err := s.mailer.SendVerificationEmail(ctx, user.Email, rawToken, user.DisplayName); err != nil {
		_ = err
	}

	return &RegisterResponse{
		UserID:      user.ID.String(),
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Message:     "registration successful. check email to verify your account.",
	}, nil
}

func (s *Service) VerifyEmail(ctx context.Context, req *VerifyEmailRequest) error {
	tokenHash := s.hasher.HashToken(req.Token)

	ev, err := s.emailRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperrors.ErrInvalidInput("invalid or expired verification token")
		}
		return fmt.Errorf("fetching verification: %w", err)
	}

	if ev.IsVerified() {
		return apperrors.ErrInvalidInput("email already verified")
	}

	if ev.IsExpired() {
		return apperrors.ErrInvalidInput("verification token has expired")
	}

	now := time.Now()
	ev.VerifiedAt = &now

	if err := s.emailRepo.MarkVerified(ctx, ev.ID); err != nil {
		return fmt.Errorf("marking verified: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, ev.UserID)
	if err != nil {
		return fmt.Errorf("fetching user: %w", err)
	}

	user.EmailVerifiedAt = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("updating user: %w", err)
	}

	return nil
}

func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperrors.ErrNotFound("user not found")
		}
		return nil, fmt.Errorf("fetching user: %w", err)
	}

	return &UserResponse{
		ID:              user.ID.String(),
		Email:           user.Email,
		DisplayName:     user.DisplayName,
		EmailVerifiedAt: user.EmailVerifiedAt,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, req *UpdateProfileRequest) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperrors.ErrNotFound("user not found")
		}
		return nil, fmt.Errorf("fetching user: %w", err)
	}

	user.DisplayName = req.DisplayName

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("updating user: %w", err)
	}

	return &UserResponse{
		ID:              user.ID.String(),
		Email:           user.Email,
		DisplayName:     user.DisplayName,
		EmailVerifiedAt: user.EmailVerifiedAt,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}, nil
}

func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, req *ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperrors.ErrNotFound("user not found")
		}
		return fmt.Errorf("fetching user: %w", err)
	}

	if err := s.hasher.Verify(req.CurrentPassword, user.PasswordHash); err != nil {
		return apperrors.ErrInvalidInput("current password is incorrect")
	}

	newHash, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		return fmt.Errorf("hashing new password: %w", err)
	}

	user.PasswordHash = newHash
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	return nil
}
