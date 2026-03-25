package service

import (
	"context"
	"strings"
	"time"

	"room-booking-service/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users UserRepository
}

func NewAuthService(users UserRepository) *AuthService { return &AuthService{users: users} }

func (s *AuthService) Register(ctx context.Context, email, password string, role models.Role) (models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !strings.Contains(email, "@") || strings.TrimSpace(password) == "" || (role != models.RoleAdmin && role != models.RoleUser) {
		return models.User{}, ErrInvalidRequest
	}
	_, err := s.users.GetByEmail(ctx, email)
	if err == nil {
		return models.User{}, ErrEmailAlreadyUsed
	}
	if err != nil && err != ErrNotFound {
		return models.User{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}
	pwd := string(hash)
	return s.users.Create(ctx, models.User{ID: uuid.NewString(), Email: email, Role: role, PasswordHash: &pwd})
}

func (s *AuthService) Login(ctx context.Context, email, password string) (models.User, error) {
	user, err := s.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return models.User{}, ErrUnauthorized
	}
	if user.PasswordHash == nil || bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)) != nil {
		return models.User{}, ErrUnauthorized
	}
	return user, nil
}

var (
	DummyAdminID = "00000000-0000-0000-0000-000000000001"
	DummyUserID  = "00000000-0000-0000-0000-000000000002"
)

func (s *AuthService) EnsureDummyUser(ctx context.Context, role models.Role) (models.User, error) {
	now := time.Now().UTC()
	switch role {
	case models.RoleAdmin:
		user := models.User{ID: DummyAdminID, Email: "admin@example.com", Role: role, CreatedAt: &now}
		return user, s.users.Upsert(ctx, user)
	case models.RoleUser:
		user := models.User{ID: DummyUserID, Email: "user@example.com", Role: role, CreatedAt: &now}
		return user, s.users.Upsert(ctx, user)
	default:
		return models.User{}, ErrInvalidRequest
	}
}
