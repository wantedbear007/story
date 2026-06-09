package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/anomalyco/story/internal/application/auth"
	"github.com/anomalyco/story/internal/application/user"
)

const sessionDir = ".story"
const sessionFile = "session.json"

// StoredSession holds the authenticated session state on disk.
type StoredSession struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	SessionID    string `json:"session_id"`
	UserID       string `json:"user_id"`
	DisplayName  string `json:"display_name"`
	Email        string `json:"email"`
}

func sessionPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, sessionDir, sessionFile), nil
}

func loadSession() (*StoredSession, error) {
	path, err := sessionPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not logged in: use 'story auth login' first")
		}
		return nil, fmt.Errorf("reading session file: %w", err)
	}

	s := &StoredSession{}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("parsing session file: %w", err)
	}

	if s.AccessToken == "" || s.RefreshToken == "" {
		return nil, fmt.Errorf("session file is invalid: please login again")
	}

	return s, nil
}

func saveSession(s *StoredSession) error {
	path, err := sessionPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating session directory: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling session: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing session file: %w", err)
	}

	return nil
}

func clearSession() error {
	path, err := sessionPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("removing session file: %w", err)
	}

	return nil
}

// saveLoginSession persists the login response to disk.
func saveLoginSession(resp *auth.LoginResponse) error {
	return saveSession(&StoredSession{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		SessionID:    resp.SessionID,
		UserID:       resp.User.ID,
		DisplayName:  resp.User.DisplayName,
		Email:        resp.User.Email,
	})
}

// saveRegisterSession persists the registration (if it returned tokens in the future).
// Currently registration doesn't return tokens, so users must login separately.
func saveRegisterSession(resp *user.RegisterResponse) error {
	return nil // registration does not create a session
}

// resolveCurrentUserID reads the stored session and returns the user ID.
func resolveCurrentUserID(deps *Dependencies) (uuid.UUID, error) {
	s, err := loadSession()
	if err != nil {
		return uuid.Nil, err
	}
	uid, err := uuid.Parse(s.UserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in session: %w", err)
	}
	return uid, nil
}

// resolveCurrentSessionID reads the stored session and returns the session ID.
func resolveCurrentSessionID(deps *Dependencies) (uuid.UUID, error) {
	s, err := loadSession()
	if err != nil {
		return uuid.Nil, err
	}
	sid, err := uuid.Parse(s.SessionID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid session ID in session: %w", err)
	}
	return sid, nil
}
