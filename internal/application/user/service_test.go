package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/anomalyco/story/internal/application/user"
	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

// mockUserRepo implements domain.UserRepository for testing.
type mockUserRepo struct {
	users map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*domain.User)}
}

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	if _, exists := m.users[u.Email]; exists {
		return domain.ErrAlreadyExists
	}
	m.users[u.Email] = u
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, exists := m.users[email]
	if !exists {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) Update(ctx context.Context, u *domain.User) error {
	if _, exists := m.users[u.Email]; !exists {
		return domain.ErrNotFound
	}
	m.users[u.Email] = u
	return nil
}

func (m *mockUserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	for _, u := range m.users {
		if u.ID == id {
			delete(m.users, u.Email)
			return nil
		}
	}
	return domain.ErrNotFound
}

// mockSessionRepo implements domain.SessionRepository for testing.
type mockSessionRepo struct {
	tokens map[string]*domain.RefreshToken
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{tokens: make(map[string]*domain.RefreshToken)}
}

func (m *mockSessionRepo) CreateRefreshToken(ctx context.Context, t *domain.RefreshToken) error {
	m.tokens[t.TokenHash] = t
	return nil
}

func (m *mockSessionRepo) GetRefreshTokenByHash(ctx context.Context, h string) (*domain.RefreshToken, error) {
	t, exists := m.tokens[h]
	if !exists {
		return nil, domain.ErrNotFound
	}
	return t, nil
}

func (m *mockSessionRepo) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockSessionRepo) RevokeUserRefreshTokens(ctx context.Context, uid uuid.UUID) error {
	return nil
}

func (m *mockSessionRepo) CreatePasswordResetToken(ctx context.Context, t *domain.PasswordResetToken) error {
	return nil
}

func (m *mockSessionRepo) GetPasswordResetTokenByHash(ctx context.Context, h string) (*domain.PasswordResetToken, error) {
	return nil, domain.ErrNotFound
}

func (m *mockSessionRepo) MarkPasswordResetTokenUsed(ctx context.Context, id uuid.UUID) error {
	return nil
}

// mockHasher implements user.PasswordHasher for testing.
type mockHasher struct{}

func (m *mockHasher) Hash(password string) (string, error) {
	return "hashed:" + password, nil
}

func (m *mockHasher) Verify(password, hash string) error {
	if hash == "hashed:"+password {
		return nil
	}
	return errors.New("password mismatch")
}

// mockTokenService implements user.TokenService for testing.
type mockTokenService struct{}

func (m *mockTokenService) GenerateAccessToken(userID uuid.UUID) (string, error) {
	return "access_token_" + userID.String(), nil
}

func (m *mockTokenService) GenerateRefreshToken() (string, error) {
	return "refresh_token_test", nil
}

func (m *mockTokenService) ValidateAccessToken(token string) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func TestService_Register(t *testing.T) {
	t.Parallel()

	svc := user.NewService(newMockUserRepo(), newMockSessionRepo(), &mockHasher{}, &mockTokenService{})

	t.Run("successful registration", func(t *testing.T) {
		resp, err := svc.Register(context.Background(), user.RegisterRequest{
			Email:       "test@example.com",
			Password:    "password123",
			DisplayName: "Test User",
		})
		if err != nil {
			t.Fatalf("Register() error = %v", err)
		}
		if resp.User.Email != "test@example.com" {
			t.Errorf("Email = %q, want %q", resp.User.Email, "test@example.com")
		}
		if resp.AccessToken == "" {
			t.Error("AccessToken should not be empty")
		}
		if resp.RefreshToken == "" {
			t.Error("RefreshToken should not be empty")
		}
	})

	t.Run("duplicate email returns error", func(t *testing.T) {
		_, err := svc.Register(context.Background(), user.RegisterRequest{
			Email:       "test@example.com",
			Password:    "password123",
			DisplayName: "Test User",
		})
		if err == nil {
			t.Fatal("Register() expected error for duplicate email")
		}
		if !errors.Is(err, domain.ErrAlreadyExists) {
			t.Errorf("Register() error = %v, want %v", err, domain.ErrAlreadyExists)
		}
	})
}

func TestService_Login(t *testing.T) {
	t.Parallel()

	svc := user.NewService(newMockUserRepo(), newMockSessionRepo(), &mockHasher{}, &mockTokenService{})

	_, _ = svc.Register(context.Background(), user.RegisterRequest{
		Email:       "login@example.com",
		Password:    "password123",
		DisplayName: "Login User",
	})

	t.Run("successful login", func(t *testing.T) {
		resp, err := svc.Login(context.Background(), user.LoginRequest{
			Email:    "login@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("Login() error = %v", err)
		}
		if resp.User.Email != "login@example.com" {
			t.Errorf("Email = %q, want %q", resp.User.Email, "login@example.com")
		}
	})

	t.Run("wrong password returns unauthorized", func(t *testing.T) {
		_, err := svc.Login(context.Background(), user.LoginRequest{
			Email:    "login@example.com",
			Password: "wrongpassword",
		})
		if err == nil {
			t.Fatal("Login() expected error for wrong password")
		}
		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Errorf("Login() error = %v, want %v", err, domain.ErrUnauthorized)
		}
	})
}

func TestService_GetProfile(t *testing.T) {
	t.Parallel()

	svc := user.NewService(newMockUserRepo(), newMockSessionRepo(), &mockHasher{}, &mockTokenService{})

	resp, _ := svc.Register(context.Background(), user.RegisterRequest{
		Email:       "profile@example.com",
		Password:    "password123",
		DisplayName: "Profile User",
	})

	t.Run("existing user returns profile", func(t *testing.T) {
		profile, err := svc.GetProfile(context.Background(), resp.User.ID)
		if err != nil {
			t.Fatalf("GetProfile() error = %v", err)
		}
		if profile.DisplayName != "Profile User" {
			t.Errorf("DisplayName = %q, want %q", profile.DisplayName, "Profile User")
		}
	})

	t.Run("non-existent user returns not found", func(t *testing.T) {
		_, err := svc.GetProfile(context.Background(), uuid.New())
		if err == nil {
			t.Fatal("GetProfile() expected error for non-existent user")
		}
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("GetProfile() error = %v, want %v", err, domain.ErrNotFound)
		}
	})
}
