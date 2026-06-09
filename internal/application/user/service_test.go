package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/anomalyco/story/internal/application/user"
	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

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

type mockEmailVerificationRepo struct {
	tokens map[string]*domain.EmailVerification
}

func newMockEmailVerificationRepo() *mockEmailVerificationRepo {
	return &mockEmailVerificationRepo{tokens: make(map[string]*domain.EmailVerification)}
}

func (m *mockEmailVerificationRepo) Create(ctx context.Context, ev *domain.EmailVerification) error {
	m.tokens[ev.TokenHash] = ev
	return nil
}

func (m *mockEmailVerificationRepo) GetByTokenHash(ctx context.Context, h string) (*domain.EmailVerification, error) {
	ev, exists := m.tokens[h]
	if !exists {
		return nil, domain.ErrNotFound
	}
	return ev, nil
}

func (m *mockEmailVerificationRepo) MarkVerified(ctx context.Context, id uuid.UUID) error {
	for _, ev := range m.tokens {
		if ev.ID == id {
			now := time.Now()
			ev.VerifiedAt = &now
			return nil
		}
	}
	return domain.ErrNotFound
}

type mockSessionRepo struct {
	sessions map[uuid.UUID]*domain.Session
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{sessions: make(map[uuid.UUID]*domain.Session)}
}

func (m *mockSessionRepo) Create(ctx context.Context, s *domain.Session) error {
	m.sessions[s.ID] = s
	return nil
}

func (m *mockSessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	s, exists := m.sessions[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	return s, nil
}

func (m *mockSessionRepo) GetByTokenHash(ctx context.Context, h string) (*domain.Session, error) {
	for _, s := range m.sessions {
		if s.TokenHash == h {
			return s, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockSessionRepo) ListByUserID(ctx context.Context, uid uuid.UUID) ([]*domain.Session, error) {
	var result []*domain.Session
	for _, s := range m.sessions {
		if s.UserID == uid {
			result = append(result, s)
		}
	}
	if result == nil {
		result = make([]*domain.Session, 0)
	}
	return result, nil
}

func (m *mockSessionRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	s, exists := m.sessions[id]
	if !exists {
		return domain.ErrNotFound
	}
	s.IsRevoked = true
	return nil
}

func (m *mockSessionRepo) RevokeAllForUser(ctx context.Context, uid uuid.UUID) error {
	for _, s := range m.sessions {
		if s.UserID == uid {
			s.IsRevoked = true
		}
	}
	return nil
}

func (m *mockSessionRepo) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	return nil
}

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

func (m *mockHasher) HashToken(token string) string {
	if token == "" {
		return "empty-hash"
	}
	return "token-hash:" + token
}

type mockMailer struct{}

func (m *mockMailer) SendVerificationEmail(ctx context.Context, to, token, displayName string) error {
	return nil
}

func TestService_Register(t *testing.T) {
	t.Parallel()

	svc := user.NewService(
		newMockUserRepo(),
		newMockEmailVerificationRepo(),
		newMockSessionRepo(),
		&mockHasher{},
		&mockMailer{},
		24*time.Hour,
	)

	t.Run("successful registration", func(t *testing.T) {
		resp, err := svc.Register(context.Background(), &user.RegisterRequest{
			Email:       "test@example.com",
			Password:    "password123",
			DisplayName: "Test User",
		})
		if err != nil {
			t.Fatalf("Register() error = %v", err)
		}
		if resp.Email != "test@example.com" {
			t.Errorf("Email = %q, want %q", resp.Email, "test@example.com")
		}
		if resp.Message == "" {
			t.Error("Message should not be empty")
		}
	})

	t.Run("duplicate email returns conflict", func(t *testing.T) {
		_, err := svc.Register(context.Background(), &user.RegisterRequest{
			Email:       "test@example.com",
			Password:    "password123",
			DisplayName: "Test User",
		})
		if err == nil {
			t.Fatal("Register() expected error for duplicate email")
		}
	})
}

func TestService_GetProfile(t *testing.T) {
	t.Parallel()

	svc := user.NewService(
		newMockUserRepo(),
		newMockEmailVerificationRepo(),
		newMockSessionRepo(),
		&mockHasher{},
		&mockMailer{},
		24*time.Hour,
	)

	resp, _ := svc.Register(context.Background(), &user.RegisterRequest{
		Email:       "profile@example.com",
		Password:    "password123",
		DisplayName: "Profile User",
	})

	uid, _ := uuid.Parse(resp.UserID)

	t.Run("existing user returns profile", func(t *testing.T) {
		profile, err := svc.GetProfile(context.Background(), uid)
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
	})
}

func TestService_ChangePassword(t *testing.T) {
	t.Parallel()

	svc := user.NewService(
		newMockUserRepo(),
		newMockEmailVerificationRepo(),
		newMockSessionRepo(),
		&mockHasher{},
		&mockMailer{},
		24*time.Hour,
	)

	resp, _ := svc.Register(context.Background(), &user.RegisterRequest{
		Email:       "changepw@example.com",
		Password:    "oldpassword",
		DisplayName: "Change PW",
	})

	uid, _ := uuid.Parse(resp.UserID)

	t.Run("successful password change", func(t *testing.T) {
		err := svc.ChangePassword(context.Background(), uid, &user.ChangePasswordRequest{
			CurrentPassword: "oldpassword",
			NewPassword:     "newpassword",
		})
		if err != nil {
			t.Fatalf("ChangePassword() error = %v", err)
		}
	})

	t.Run("wrong current password returns error", func(t *testing.T) {
		err := svc.ChangePassword(context.Background(), uid, &user.ChangePasswordRequest{
			CurrentPassword: "wrongpassword",
			NewPassword:     "newpassword",
		})
		if err == nil {
			t.Fatal("ChangePassword() expected error for wrong password")
		}
	})
}
