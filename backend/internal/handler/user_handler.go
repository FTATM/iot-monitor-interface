package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type UserHandler struct {
	service model.UserService
}

func NewUserHandler(service model.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var res Response
	var createUserReq model.CreateUserRequest
	var err error
	if err = json.NewDecoder(r.Body).Decode(&createUserReq); err != nil {
		res.Message = "Invalid request body"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if createUserReq.Username == "" || createUserReq.Password == "" {
		res.Message = "Username and password are required"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if createUserReq.FirstName == "" || createUserReq.LastName == "" {
		res.Message = "FirstName and LastName are required"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	user, err := h.service.Create(r.Context(), &createUserReq)

	res.Data = map[string]any{
		"userId":   user.UserId,
		"username": user.Username,
		"active":   user.Active,
	}
	respondJson(w, http.StatusCreated, &res)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var res Response
	var creds model.LoginCredentials
	var err error
	if err = json.NewDecoder(r.Body).Decode(&creds); err != nil {
		res.Message = "Invalid request body"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	issueTime := time.Now()
	expTime := time.Now().Add(24 * time.Hour)
	user, tokenString, err := h.service.LoginUserJwt(r.Context(), &creds, issueTime, expTime)
	if err != nil {
		res.Message = "Invalid user"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    tokenString,
		Expires:  expTime,
		HttpOnly: true,                    // CRITICAL: Prevents JavaScript/XSS from reading the cookie
		Secure:   true,                    // CRITICAL: Ensures cookie is only sent over HTTPS (set to false ONLY if testing on localhost HTTP)
		SameSite: http.SameSiteStrictMode, // Protects against Cross-Site Request Forgery (CSRF)
		Path:     "/",
	})

	res.Data = map[string]any{
		"user": map[string]any{
			"id":        user.UserId,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
		},
	}
	respondJson(w, http.StatusOK, &res)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    "",
		Expires:  time.Unix(0, 0), // Set date to the past to delete it
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}
