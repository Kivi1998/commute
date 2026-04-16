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
var DefaultSeedUsers = []SeedUser{
	{Email: "jiahao@diit.cn", Name: "贾昊", Password: "commute123"},
	{Email: "demo@example.com", Name: "Demo", Password: "demo123"},
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
			// 主账号：占用 id=1（已有默认用户），更新其 email/name/password
			if _, err := pool.Exec(ctx, `
                UPDATE app_user SET
                    email = $1, name = $2, password_hash = $3, updated_at = NOW()
                WHERE id = 1 AND (password_hash IS NULL OR password_hash = '')
            `, s.Email, s.Name, hash); err != nil {
				return fmt.Errorf("update primary seed user: %w", err)
			}
			// 若 id=1 已被其他 email 占用（历史遗留），保底插入一个新用户
			var exists bool
			if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM app_user WHERE email = $1)`, s.Email).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				if _, err := pool.Exec(ctx, `
                    INSERT INTO app_user (email, name, password_hash) VALUES ($1, $2, $3)
                    ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash
                `, s.Email, s.Name, hash); err != nil {
					return fmt.Errorf("insert primary seed user: %w", err)
				}
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
