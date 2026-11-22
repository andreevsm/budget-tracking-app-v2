package service

import (
	"budget/internal/domain"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) issueTokens(ctx context.Context, userID uuid.UUID) (domain.Tokens, error) {
	claims := jwt.MapClaims{
		"sub":  userID.String(),
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(s.conf.AccessTTL).Unix(),
		"type": "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	access, err := token.SignedString([]byte(s.conf.JWTSecret))

	if err != nil {
		return domain.Tokens{}, err
	}

	buf := make([]byte, 32)

	if _, err := rand.Read(buf); err != nil {
		return domain.Tokens{}, err
	}

	refresh := hex.EncodeToString(buf)

	sum := sha256.Sum256([]byte(refresh))

	if err := s.repo.CreateSession(ctx, userID, hex.EncodeToString(sum[:]), time.Now().Add(s.conf.RefreshTTL)); err != nil {
		return domain.Tokens{}, err
	}

	return domain.Tokens{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *Service) Register(ctx context.Context, in domain.RegisterInput) (domain.User, domain.Tokens, error) {
	addr, err := mail.ParseAddress(in.Email)

	if err != nil {
		return domain.User{}, domain.Tokens{}, fmt.Errorf("register user: %w: %v", domain.ErrInvalidInput, err)
	}

	if len(in.Password) < 8 {
		return domain.User{}, domain.Tokens{}, fmt.Errorf("register user: %w: %v", domain.ErrInvalidInput, "password must be > 8")
	}

	if in.DisplayName == "" {
		return domain.User{}, domain.Tokens{}, fmt.Errorf("register user: %w: %v", domain.ErrInvalidInput, "display_name is empty")
	}

	in.Email = strings.ToLower(strings.TrimSpace(addr.Address))

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)

	if err != nil {
		return domain.User{}, domain.Tokens{}, fmt.Errorf("register user: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, in, string(passwordHash))

	if err != nil {
		return domain.User{}, domain.Tokens{}, fmt.Errorf("register user: %w", err)
	}

	tokens, err := s.issueTokens(ctx, user.ID)
	if err != nil {
		return domain.User{}, domain.Tokens{}, fmt.Errorf("register user: %w", err)
	}

	return user, tokens, nil

}

func (s *Service) Login(ctx context.Context, in domain.LoginInput) (domain.User, domain.Tokens, error) {
	addr, err := mail.ParseAddress(in.Email)

	if err != nil {
		return domain.User{}, domain.Tokens{}, fmt.Errorf("login user: %w: %v", domain.ErrInvalidInput, err)
	}

	in.Email = strings.ToLower(strings.TrimSpace(addr.Address))

	user, passwordHash, err := s.repo.GetUserByEmail(ctx, in.Email)

	if err != nil {
		return domain.User{}, domain.Tokens{}, fmt.Errorf("login user: %w", domain.ErrInvalidCredentials)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(in.Password)); err != nil {
		return domain.User{}, domain.Tokens{}, fmt.Errorf("login user: %w", domain.ErrInvalidCredentials)
	}

	tokens, err := s.issueTokens(ctx, user.ID)

	if err != nil {
		return domain.User{}, domain.Tokens{}, fmt.Errorf("login user: %w", err)
	}

	return user, tokens, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (domain.Tokens, error) {
	sum := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(sum[:])

	sess, err := s.repo.GetSessionByHash(ctx, tokenHash)

	if err != nil {
		return domain.Tokens{}, fmt.Errorf("refresh: %w", domain.ErrUnauthorized)
	}

	if sess.RevokedAt != nil || time.Now().After(sess.ExpiresAt) {
		return domain.Tokens{}, fmt.Errorf("refresh: %w", domain.ErrUnauthorized)
	}

	if err := s.repo.RevokeSession(ctx, tokenHash); err != nil {
		return domain.Tokens{}, fmt.Errorf("refresh: %w", err)
	}

	tokens, err := s.issueTokens(ctx, sess.UserID)
	if err != nil {
		return domain.Tokens{}, fmt.Errorf("refresh: %w", err)
	}

	return tokens, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	sum := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(sum[:])

	if err := s.repo.RevokeSession(ctx, tokenHash); err != nil {
		return fmt.Errorf("logout: %w", err)
	}

	return nil
}
