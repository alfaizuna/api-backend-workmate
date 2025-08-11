package auth

import (
	"context"
	"errors"
	"time"

	"backend-work-mate/internal/config"
	"backend-work-mate/internal/storage/postgres"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	users     postgres.UserRepository
	jwtSecret []byte
}

func NewService(repo postgres.UserRepository, cfg *config.Config) *Service {
	return &Service{
		users:     repo,
		jwtSecret: []byte(cfg.JWTSecret),
	}
}

type RegisterInput struct {
	Name       string  `json:"name" binding:"required"`
	Email      string  `json:"email" binding:"required,email"`
	Password   string  `json:"password" binding:"required,min=6"`
	Department *string `json:"department"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthToken struct {
	Token string `json:"token"`
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*postgres.User, error) {
	existing, err := s.users.GetByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email sudah terdaftar")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &postgres.User{
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: string(hash),
		Role:         postgres.RoleEmployee,
		Department:   in.Department,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (*AuthToken, error) {
	user, err := s.users.GetByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("email atau password salah")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return nil, errors.New("email atau password salah")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	})
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}
	return &AuthToken{Token: signed}, nil
}
