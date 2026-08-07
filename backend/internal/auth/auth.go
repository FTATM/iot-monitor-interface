package auth

type ContextKey string

const (
	CookieName               = "authToken"
	AuthUserIdKey ContextKey = "userId"
)
