package auth

import "github.com/golang-jwt/jwt/v5"

type Claim struct {
	UserId    int    `json:"userId"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	// Permissions map[string][]string `json:"permissions"`
	jwt.RegisteredClaims
}
