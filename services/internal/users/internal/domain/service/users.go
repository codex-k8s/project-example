package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/codex-k8s/project-example/services/internal/users/internal/domain/errs"
	userrepo "github.com/codex-k8s/project-example/services/internal/users/internal/domain/repository/user"
	"github.com/codex-k8s/project-example/services/internal/users/internal/domain/types/entity"
	"golang.org/x/crypto/bcrypt"
)

// Service implements users use-cases.
type Service struct {
	repo userrepo.Repository
}

// New constructs Service.
func New(repo userrepo.Repository) *Service {
	return &Service{repo: repo}
}

// Register creates a new user with a bcrypt-hashed password.
func (s *Service) Register(ctx context.Context, username, password string) (entity.User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return entity.User{}, errs.Validation{Field: "username", Msg: "required"}
	}
	if len(username) > 64 {
		return entity.User{}, errs.Validation{Field: "username", Msg: "too long"}
	}
	if password == "" {
		return entity.User{}, errs.Validation{Field: "password", Msg: "required"}
	}
	if len(password) < 8 {
		return entity.User{}, errs.Validation{Field: "password", Msg: "too short"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return entity.User{}, fmt.Errorf("hash password: %w", err)
	}

	u, err := s.repo.Create(ctx, entity.User{
		Username:     username,
		PasswordHash: string(hash),
	})
	if err != nil {
		// Repo may return Conflict; bubble it up as-is.
		var c errs.Conflict
		if errors.As(err, &c) {
			return entity.User{}, c
		}
		return entity.User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// Authenticate verifies credentials and returns a user on success.
func (s *Service) Authenticate(ctx context.Context, username, password string) (entity.User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return entity.User{}, errs.Validation{Field: "username", Msg: "required"}
	}
	if password == "" {
		return entity.User{}, errs.Validation{Field: "password", Msg: "required"}
	}

	u, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		var nf errs.NotFound
		if errors.As(err, &nf) {
			// Do not reveal whether the user exists.
			return entity.User{}, errs.Unauthorized{Msg: "invalid credentials"}
		}
		return entity.User{}, fmt.Errorf("get user by username: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return entity.User{}, errs.Unauthorized{Msg: "invalid credentials"}
	}
	return u, nil
}

// GetUser returns a user by ID.
func (s *Service) GetUser(ctx context.Context, id int64) (entity.User, error) {
	if id <= 0 {
		return entity.User{}, errs.Validation{Field: "id", Msg: "invalid"}
	}
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return entity.User{}, err
	}
	return u, nil
}
