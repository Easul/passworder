package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"

	"passworder/internal/auth"
	"passworder/internal/model"
	"passworder/internal/repository"
	"passworder/internal/service"
)

type AuthHandler struct {
	authService    *auth.Service
	authRepo       *repository.AuthRepository
	settingService *service.SettingService
}

func NewAuthHandler(authService *auth.Service, authRepo *repository.AuthRepository, settingService *service.SettingService) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		authRepo:       authRepo,
		settingService: settingService,
	}
}

type SetupRequest struct {
	Password string `json:"password"`
}

type LoginRequest struct {
	Password string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Setup(w http.ResponseWriter, r *http.Request) {
	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if h.authService.IsInitialized() {
		JSONError(w, http.StatusConflict, "Already initialized")
		return
	}

	if err := h.authService.Setup(req.Password); err != nil {
		JSONError(w, http.StatusInternalServerError, "Setup failed")
		return
	}

	now := model.Now()
	authRecord := &model.Auth{
		ID:           1,
		PasswordHash: h.authService.GetPasswordHash(),
		KDFSalt:      h.authService.GetSalt(),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := h.authRepo.Save(authRecord); err != nil {
		JSONError(w, http.StatusInternalServerError, "Save failed")
		return
	}

	session, err := h.authService.CreateSession()
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Session creation failed")
		return
	}

	JSONResponse(w, http.StatusOK, TokenResponse{Token: session.Token})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if err := h.authService.Verify(req.Password); err != nil {
		JSONError(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	session, err := h.authService.CreateSession()
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Session creation failed")
		return
	}

	JSONResponse(w, http.StatusOK, TokenResponse{Token: session.Token})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.authService.DestroySession()
	JSONResponse(w, http.StatusOK, nil)
}

func (h *AuthHandler) Check(w http.ResponseWriter, r *http.Request) {
	initialized := h.authService.IsInitialized()
	JSONResponse(w, http.StatusOK, map[string]bool{"initialized": initialized})
}

func PasswordGenerator(w http.ResponseWriter, r *http.Request) {
	length := 16
	if l := r.URL.Query().Get("length"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &length); err != nil {
			length = 16
		}
	}
	if length < 8 {
		length = 8
	}
	if length > 64 {
		length = 64
	}

	includeUpper := r.URL.Query().Get("upper") != "false"
	includeLower := r.URL.Query().Get("lower") != "false"
	includeNumbers := r.URL.Query().Get("numbers") != "false"
	includeSymbols := r.URL.Query().Get("symbols") != "false"

	password := generatePassword(length, includeUpper, includeLower, includeNumbers, includeSymbols)
	JSONResponse(w, http.StatusOK, map[string]string{"password": password})
}

func generatePassword(length int, upper, lower, numbers, symbols bool) string {
	var charset string
	if upper {
		charset += "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	}
	if lower {
		charset += "abcdefghijklmnopqrstuvwxyz"
	}
	if numbers {
		charset += "0123456789"
	}
	if symbols {
		charset += "!@#$%^&*()_+-=[]{}|;:,.<>?"
	}
	if charset == "" {
		charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	}

	result := make([]byte, length)
	for i := range result {
		b := make([]byte, 1)
		rand.Read(b)
		result[i] = charset[int(b[0])%len(charset)]
	}
	return string(result)
}
