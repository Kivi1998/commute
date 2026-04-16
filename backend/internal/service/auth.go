package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/haojia/commute/internal/config"
	"github.com/haojia/commute/internal/model"
	"github.com/haojia/commute/internal/pkg/auth"
	"github.com/haojia/commute/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type AuthService struct {
	repo *repository.UserRepo
	cfg  config.AuthConfig
}

func NewAuthService(repo *repository.UserRepo, cfg config.AuthConfig) *AuthService {
	return &AuthService{repo: repo, cfg: cfg}
}

func (s *AuthService) Login(ctx context.Context, in model.LoginInput) (*model.LoginResponse, error) {
	user, hash, err := s.repo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(in.Email)))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if hash == "" || !auth.VerifyPassword(hash, in.Password) {
		return nil, ErrInvalidCredentials
	}

	ttl := s.cfg.TokenTTL()
	if ttl <= 0 {
		ttl = 7 * 24 * time.Hour
	}
	token, err := auth.Issue(s.cfg.JWTSecret, user.ID, user.Email, ttl)
	if err != nil {
		return nil, err
	}
	return &model.LoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(ttl),
		User:      *user,
	}, nil
}

func (s *AuthService) GetUser(ctx context.Context, id int64) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}
