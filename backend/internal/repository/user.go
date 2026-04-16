package repository

import (
	"context"
	"errors"

	"github.com/haojia/commute/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

type UserRecord struct {
	ID           int64
	Name         *string
	Email        *string
	Phone        *string
	PasswordHash *string
	CreatedAt    string
}

// GetByEmail 根据邮箱查找（未找到返回 ErrNotFound）
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, string, error) {
	const q = `
        SELECT id, COALESCE(name, ''), COALESCE(email, ''), COALESCE(phone, ''),
               COALESCE(password_hash, ''), created_at
        FROM app_user
        WHERE email = $1 AND deleted_at IS NULL
    `
	var u model.User
	var hash string
	err := r.pool.QueryRow(ctx, q, email).Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &hash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", ErrNotFound
	}
	if err != nil {
		return nil, "", err
	}
	return &u, hash, nil
}

// GetByID 主要用于中间件校验 token 有效后补齐用户信息
func (r *UserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	const q = `
        SELECT id, COALESCE(name, ''), COALESCE(email, ''), COALESCE(phone, ''), created_at
        FROM app_user
        WHERE id = $1 AND deleted_at IS NULL
    `
	var u model.User
	err := r.pool.QueryRow(ctx, q, id).Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
