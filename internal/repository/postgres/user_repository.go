package postgres

import (
	"context"
	"errors"

	"room-booking-service/internal/models"
	"room-booking-service/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) Create(ctx context.Context, user models.User) (models.User, error) {
	var created models.User
	var pwd *string
	err := r.db.QueryRow(ctx, `INSERT INTO users (id,email,password_hash,role) VALUES ($1,$2,$3,$4) RETURNING id,email,password_hash,role,created_at`, user.ID, user.Email, user.PasswordHash, user.Role).
		Scan(&created.ID, &created.Email, &pwd, &created.Role, &created.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.User{}, service.ErrEmailAlreadyUsed
		}
		return models.User{}, err
	}
	created.PasswordHash = pwd
	return created, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	var pwd *string
	err := r.db.QueryRow(ctx, `SELECT id,email,password_hash,role,created_at FROM users WHERE email=$1`, email).
		Scan(&user.ID, &user.Email, &pwd, &user.Role, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, service.ErrNotFound
	}
	user.PasswordHash = pwd
	return user, err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (models.User, error) {
	var user models.User
	var pwd *string
	err := r.db.QueryRow(ctx, `SELECT id,email,password_hash,role,created_at FROM users WHERE id=$1`, id).
		Scan(&user.ID, &user.Email, &pwd, &user.Role, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, service.ErrNotFound
	}
	user.PasswordHash = pwd
	return user, err
}

func (r *UserRepository) Upsert(ctx context.Context, user models.User) error {
	_, err := r.db.Exec(ctx, `INSERT INTO users (id,email,role) VALUES ($1,$2,$3) ON CONFLICT (id) DO UPDATE SET email=EXCLUDED.email, role=EXCLUDED.role`, user.ID, user.Email, user.Role)
	return err
}
