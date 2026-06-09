package user

import (
	"time"

	"github.com/anomalyco/story/internal/domain"
	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8,max=128"`
	DisplayName string `json:"display_name" validate:"required,min=1,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name" validate:"omitempty,min=1,max=100"`
}

type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

func UserToResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt,
	}
}
