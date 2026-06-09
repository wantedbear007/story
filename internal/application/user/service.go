package user

import (
	"context"
	"fmt"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

// PasswordHasher abstracts password hashing so the service layer
// never depends on a specific algorithm (Argon2, bcrypt, etc.).
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}

// TokenService abstracts JWT token generation so the service layer
// never depends on a specific JWT implementation.
type TokenService interface {
	GenerateAccessToken(userID uuid.UUID) (string, error)
	GenerateRefreshToken() (string, error)
	ValidateAccessToken(tokenString string) (uuid.UUID, error)
}

// Service implements user-related use cases.
// It depends on interfaces, not concrete implementations (DIP).
type Service struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
	hasher      PasswordHasher
	tokens      TokenService
}

func NewService(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	hasher PasswordHasher,
	tokens TokenService,
) *Service {
	return &Service{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		hasher:      hasher,
		tokens:      tokens,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && err != domain.ErrNotFound {
		return nil, fmt.Errorf("checking existing user: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("%w: email already registered", domain.ErrAlreadyExists)
	}

	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hash,
		DisplayName:  req.DisplayName,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	accessToken, err := s.tokens.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	rawRefresh, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	refreshHash, err := s.hasher.Hash(rawRefresh)
	if err != nil {
		return nil, fmt.Errorf("hashing refresh token: %w", err)
	}

	refreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: refreshHash,
	}

	if err := s.sessionRepo.CreateRefreshToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		User:         UserToResponse(user),
	}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid email or password", domain.ErrUnauthorized)
	}

	if err := s.hasher.Verify(req.Password, user.PasswordHash); err != nil {
		return nil, fmt.Errorf("%w: invalid email or password", domain.ErrUnauthorized)
	}

	accessToken, err := s.tokens.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	rawRefresh, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	refreshHash, err := s.hasher.Hash(rawRefresh)
	if err != nil {
		return nil, fmt.Errorf("hashing refresh token: %w", err)
	}

	refreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: refreshHash,
	}

	if err := s.sessionRepo.CreateRefreshToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		User:         UserToResponse(user),
	}, nil
}

func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: user not found", domain.ErrNotFound)
	}

	resp := UserToResponse(user)
	return &resp, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: user not found", domain.ErrNotFound)
	}

	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("updating user: %w", err)
	}

	resp := UserToResponse(user)
	return &resp, nil
}
