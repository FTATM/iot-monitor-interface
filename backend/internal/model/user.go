package model

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	UserId       int        `json:"userId" db:"user_id"`
	FirstName    string     `json:"firstName" db:"first_name"`
	LastName     string     `json:"lastName" db:"last_name"`
	Username     string     `json:"username" db:"username"`
	PasswordHash string     `json:"password_hash" db:"password_hash"`
	Active       bool       `json:"active" db:"active"`
	CreatedAt    time.Time  `json:"createAt" db:"created_at"`
	UpdatedAt    *time.Time `json:"updateAt" db:"updated_at"`
}

type UserDetail struct {
	UserId    int    `json:"userId"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	// Widgets []Widget `json:"widgets"`
}

type LoginCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	UserID   int    `json:"userId"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type CreateUserRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Active    bool   `json:"active"`
}

type UserRepository interface {
	GetById(ctx context.Context, id int) (*User, error)
	Create(ctx context.Context, user *User) error
	GetByUsername(ctx context.Context, username string) (*User, error)
	UserCount(ctx context.Context) (int, error)
	// CreateCanvas(ctx context.Context) (*Canvas, error)
}

type UserService interface {
	LoginUserJwt(ctx context.Context, creds *LoginCredentials, issueTime time.Time, expTime time.Time) (*User, string, error)
	Create(ctx context.Context, createUserReq *CreateUserRequest) (*User, error)
	// GetUserDetail(ctx context.Context, id int) (*CanvasDetail, error)
	// CreateCanvas(ctx context.Context) (*Canvas, error)
}
