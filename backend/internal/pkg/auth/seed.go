package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedUser 预置账号
type SeedUser struct {
	Email    string
	Name     string
	Password string
}

// DefaultSeedUsers 项目默认账号（MVP 固定账号模式，无注册）
//
// Email 字段作为"登录账号名"使用，不强制邮箱格式。
var DefaultSeedUsers = []SeedUser{
	{Email: "kivi", Name: "Kivi", Password: "542426"},
	{Email: "dudu", Name: "Dudu", Password: "311416"},
}

// EnsureSeedUsers 启动时确保固定账号存在
//
// 策略：
// - 主账号 (seeds[0]) 对应 user_id=1（把现有的默认用户升级为该账号，保留已有通勤数据）
// - 其他账号：若 email 不存在则 INSERT 新用户
func EnsureSeedUsers(ctx context.Context, pool *pgxpool.Pool, seeds []SeedUser) error {
	if len(seeds) == 0 {
		return nil
	}
	for i, s := range seeds {
		hash, err := HashPassword(s.Password)
		if err != nil {
			return fmt.Errorf("hash password for %s: %w", s.Email, err)
		}

		if i == 0 {
			// 主账号：强制绑定 id=1（历史 commute 数据都挂在 user_id=1 下）
			// 先清空其他行上同名 email 避免唯一冲突
			if _, err := pool.Exec(ctx,
				`UPDATE app_user SET email = NULL WHERE email = $1 AND id <> 1`, s.Email,
			); err != nil {
				return fmt.Errorf("clear other rows with same email: %w", err)
			}
			if _, err := pool.Exec(ctx, `
                UPDATE app_user SET
                    email = $1, name = $2, password_hash = $3, updated_at = NOW()
                WHERE id = 1
            `, s.Email, s.Name, hash); err != nil {
				return fmt.Errorf("update primary seed user: %w", err)
			}
			continue
		}

		if _, err := pool.Exec(ctx, `
            INSERT INTO app_user (email, name, password_hash)
            VALUES ($1, $2, $3)
            ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash,
                                               name = EXCLUDED.name,
                                               updated_at = NOW()
        `, s.Email, s.Name, hash); err != nil {
			return fmt.Errorf("seed user %s: %w", s.Email, err)
		}
	}
	return nil
}
