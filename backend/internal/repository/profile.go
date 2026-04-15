package repository

import (
	"context"
	"errors"

	"github.com/haojia/commute/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type ProfileRepo struct {
	pool *pgxpool.Pool
}

func NewProfileRepo(pool *pgxpool.Pool) *ProfileRepo {
	return &ProfileRepo{pool: pool}
}

func (r *ProfileRepo) Get(ctx context.Context, userID int64) (*model.Profile, error) {
	const q = `
        SELECT id, user_id, current_city, current_city_code, target_position,
               experience_years, COALESCE(preferred_company_types::text[], '{}'),
               created_at, updated_at
        FROM user_profile WHERE user_id = $1
    `
	p := &model.Profile{}
	err := r.pool.QueryRow(ctx, q, userID).Scan(
		&p.ID, &p.UserID, &p.CurrentCity, &p.CurrentCityCode, &p.TargetPosition,
		&p.ExperienceYears, &p.PreferredCompanyTypes,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ProfileRepo) Upsert(ctx context.Context, userID int64, in model.ProfileUpsertInput) (*model.Profile, error) {
	const q = `
        INSERT INTO user_profile (
            user_id, current_city, current_city_code, target_position,
            experience_years, preferred_company_types
        )
        VALUES ($1, $2, $3, $4, $5, $6::company_type_enum[])
        ON CONFLICT (user_id) DO UPDATE SET
            current_city = EXCLUDED.current_city,
            current_city_code = EXCLUDED.current_city_code,
            target_position = EXCLUDED.target_position,
            experience_years = EXCLUDED.experience_years,
            preferred_company_types = EXCLUDED.preferred_company_types,
            updated_at = NOW()
        RETURNING id, user_id, current_city, current_city_code, target_position,
                  experience_years, COALESCE(preferred_company_types::text[], '{}'),
                  created_at, updated_at
    `
	types := in.PreferredCompanyTypes
	if types == nil {
		types = []string{}
	}
	p := &model.Profile{}
	err := r.pool.QueryRow(ctx, q,
		userID, in.CurrentCity, in.CurrentCityCode, in.TargetPosition,
		in.ExperienceYears, types,
	).Scan(
		&p.ID, &p.UserID, &p.CurrentCity, &p.CurrentCityCode, &p.TargetPosition,
		&p.ExperienceYears, &p.PreferredCompanyTypes,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}
