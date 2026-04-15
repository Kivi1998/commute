package repository

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/haojia/commute/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AIRepo struct {
	pool *pgxpool.Pool
}

func NewAIRepo(pool *pgxpool.Pool) *AIRepo {
	return &AIRepo{pool: pool}
}

// BuildCacheKey 根据输入生成稳定的缓存键
func BuildCacheKey(userID int64, in model.AIRecommendInput) string {
	types := append([]string{}, in.CompanyTypes...)
	sort.Strings(types)
	expY := 0
	if in.ExperienceYears != nil {
		expY = *in.ExperienceYears
	}
	raw := strings.Join([]string{
		strconv.FormatInt(userID, 10),
		in.City, in.Position,
		strconv.Itoa(expY),
		strings.Join(types, ","),
	}, "|")
	h := md5.Sum([]byte(raw))
	return hex.EncodeToString(h[:])
}

// CachedRecommendation 缓存命中的存储结构
type CachedRecommendation struct {
	ID          int64
	RawResponse json.RawMessage
	RequestedAt time.Time
	ExpiresAt   time.Time
	TokenInput  int
	TokenOutput int
}

// FindCache 查找 24h 内未过期的缓存（以 cache_key 命中）
func (r *AIRepo) FindCache(ctx context.Context, cacheKey string) (*CachedRecommendation, error) {
	const q = `
        SELECT id, raw_response, requested_at, expires_at,
               COALESCE(token_input, 0), COALESCE(token_output, 0)
        FROM ai_recommendation_cache
        WHERE cache_key = $1 AND expires_at > NOW()
        ORDER BY requested_at DESC LIMIT 1
    `
	c := &CachedRecommendation{}
	err := r.pool.QueryRow(ctx, q, cacheKey).Scan(
		&c.ID, &c.RawResponse, &c.RequestedAt, &c.ExpiresAt,
		&c.TokenInput, &c.TokenOutput,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

// InsertCache 写入新缓存
func (r *AIRepo) InsertCache(ctx context.Context, userID int64, in model.AIRecommendInput, cacheKey string,
	companies []model.AIRecommendedCompany, tokenIn, tokenOut int, ttl time.Duration,
) (*CachedRecommendation, error) {
	raw, err := json.Marshal(map[string]any{"companies": companies})
	if err != nil {
		return nil, err
	}

	types := in.CompanyTypes
	if types == nil {
		types = []string{}
	}
	const q = `
        INSERT INTO ai_recommendation_cache (
            user_id, city, position, experience_years, company_types,
            cache_key, raw_response, company_count, expires_at,
            token_input, token_output
        )
        VALUES ($1, $2, $3, $4, $5::company_type_enum[],
                $6, $7, $8, $9, $10, $11)
        RETURNING id, raw_response, requested_at, expires_at,
                  COALESCE(token_input, 0), COALESCE(token_output, 0)
    `
	c := &CachedRecommendation{}
	err = r.pool.QueryRow(ctx, q,
		userID, in.City, in.Position, in.ExperienceYears, types,
		cacheKey, raw, len(companies), time.Now().Add(ttl),
		tokenIn, tokenOut,
	).Scan(
		&c.ID, &c.RawResponse, &c.RequestedAt, &c.ExpiresAt,
		&c.TokenInput, &c.TokenOutput,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}
