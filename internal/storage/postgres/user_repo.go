package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRole string

const (
	RoleAdmin    UserRole = "Admin"
	RoleEmployee UserRole = "Employee"
)

type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	Role         UserRole
	Department   *string
	CreatedAt    time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) Create(ctx context.Context, user *User) error {
	query := `
        insert into public.users (name, email, password_hash, role, department)
        values ($1, $2, $3, $4, $5)
        returning id, created_at
    `

	return r.pool.QueryRow(ctx, query,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.Department,
	).Scan(&user.ID, &user.CreatedAt)
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
        select id, name, email, password_hash, role, department, created_at
        from public.users where email = $1 limit 1
    `

	row := r.pool.QueryRow(ctx, query, email)
	var u User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Department, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// end
