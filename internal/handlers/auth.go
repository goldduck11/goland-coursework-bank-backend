package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"awesomeProject/internal/repository/postgres"
	"awesomeProject/internal/service"

	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	service *service.AuthService
	logger  *logrus.Logger
}

func NewAuthHandler(s *service.AuthService, l *logrus.Logger) *AuthHandler {
	return &AuthHandler{service: s, logger: l}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	id, err := h.service.Register(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, postgres.ErrUserAlreadyExists) {
			respondWithError(w, http.StatusConflict, err.Error())
			return
		}
		h.logger.WithError(err).Error("Registration failed")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"user_id": id, "status": "success"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	token, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"token": token})
}
