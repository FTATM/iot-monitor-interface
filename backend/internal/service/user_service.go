package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
)

type userService struct {
	txManager model.TransactionManager
	userRepo  model.UserRepository
	jwtKey    []byte
}

func NewUserService(txManager model.TransactionManager, u model.UserRepository, jwtKey []byte) model.UserService {
	return &userService{txManager: txManager, userRepo: u, jwtKey: jwtKey}
}

func (s *userService) Create(ctx context.Context, createUserReq *model.CreateUserRequest) (*model.User, error) {
	hash, err := argon2id.CreateHash(createUserReq.Password, argon2id.DefaultParams)
	if err != nil {
		return nil, err
	}
	user := model.User{
		FirstName:    createUserReq.FirstName,
		LastName:     createUserReq.LastName,
		Username:     createUserReq.Username,
		Active:       createUserReq.Active,
		PasswordHash: hash,
	}

	err = s.userRepo.Create(ctx, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *userService) LoginUserJwt(ctx context.Context, creds *model.LoginCredentials, issueTime time.Time, expTime time.Time) (*model.User, string, error) {
	user, err := s.userRepo.GetByUsername(ctx, creds.Username)
	if err != nil {
		return nil, "", err
	}

	if !user.Active {
		return nil, "", errors.New("User not active")
	}

	match, err := argon2id.ComparePasswordAndHash(creds.Password, user.PasswordHash)
	if err != nil {
		return nil, "", errors.New("Internal server error")
	}

	if !match {
		return nil, "", errors.New("Invalid username or password")
	}

	// expirationTime := time.Now().Add(24 * time.Hour)

	claims := &model.Claims{
		UserID:   user.UserId,
		Username: creds.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
			IssuedAt:  jwt.NewNumericDate(issueTime),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Create the JWT string
	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		log.Printf("JWT generation error: %v", err)
		return nil, "", errors.New("Internal server error")
	}

	return user, tokenString, nil
}
