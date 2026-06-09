package auth

import "time"

type LoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	DeviceInfo string `json:"device_info,omitempty"`
	IPAddress  string `json:"ip_address,omitempty"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	SessionID    string       `json:"session_id"`
	User         *UserInfo    `json:"user"`
	ExpiresIn    int64        `json:"expires_in"`
}

type UserInfo struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	SessionID    string `json:"session_id"`
	ExpiresIn    int64  `json:"expires_in"`
}

type SessionResponse struct {
	ID         string     `json:"id"`
	DeviceInfo string     `json:"device_info,omitempty"`
	IPAddress  string     `json:"ip_address,omitempty"`
	IsCurrent  bool       `json:"is_current"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  time.Time  `json:"expires_at"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}
