package handler

import (
	"budget/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ctxKey string

const userIDKey ctxKey = "userID"

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var in domain.RegisterInput

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&in)

	if err != nil {
		h.logger.Warn("failed to decode user", "error", err)
		http.Error(w, "failed to decode user", http.StatusBadRequest)
		return
	}

	user, tokens, err := h.serv.Register(r.Context(), in)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			h.logger.Warn("wrong input", "error", err)
			http.Error(w, "wrong input", http.StatusBadRequest)

		case errors.Is(err, domain.ErrEmailTaken):
			h.logger.Warn("email already exists", "error", err)
			http.Error(w, "email already exists", http.StatusConflict)

		default:
			h.logger.Error("create user failed", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		return
	}

	h.logger.Info("user created", "id", user.ID)
	h.writeJSON(w, http.StatusCreated, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var in domain.LoginInput

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&in)

	if err != nil {
		h.logger.Warn("failed to decode user", "error", err)
		http.Error(w, "failed to decode user", http.StatusBadRequest)
		return
	}

	user, tokens, err := h.serv.Login(r.Context(), in)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			h.logger.Warn("wrong input", "error", err)
			http.Error(w, "wrong input", http.StatusBadRequest)

		case errors.Is(err, domain.ErrInvalidCredentials):
			h.logger.Warn("invalid credentials", "error", err)
			http.Error(w, "invalid credentials", http.StatusUnauthorized)

		default:
			h.logger.Error("login failed", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		return
	}

	h.logger.Info("user logged in", "id", user.ID)
	h.writeJSON(w, http.StatusOK, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var in domain.RefreshInput

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&in)

	if err != nil {
		h.logger.Warn("failed to decode user", "error", err)
		http.Error(w, "failed to decode user", http.StatusBadRequest)
		return
	}

	tokens, err := h.serv.Refresh(r.Context(), in.RefreshToken)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUnauthorized):
			h.logger.Warn("unauthorized", "error", err)
			http.Error(w, "unauthorized", http.StatusUnauthorized)

		default:
			h.logger.Error("refresh failed", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		return
	}

	h.logger.Info("tokens refreshed")
	h.writeJSON(w, http.StatusOK, tokens)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	var in domain.LogoutInput

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&in)

	if err != nil {
		h.logger.Warn("failed to decode user", "error", err)
		http.Error(w, "failed to decode user", http.StatusBadRequest)
		return
	}

	err = h.serv.Logout(r.Context(), in.RefreshToken)

	if err != nil {
		h.logger.Error("logout failed", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("user logout")
	w.WriteHeader(http.StatusNoContent)
}

func requireAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			tokenStr := r.Header.Get("Authorization")

			tokenArr := strings.Split(tokenStr, " ")

			if len(tokenArr) != 2 || tokenArr[0] != "Bearer" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(tokenArr[1], func(t *jwt.Token) (any, error) {
				return []byte(secret), nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

			if err != nil || token == nil || !token.Valid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims := token.Claims.(jwt.MapClaims)

			if claims["type"] != "access" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			sub, ok := claims["sub"].(string)

			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := uuid.Parse(sub)

			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}
