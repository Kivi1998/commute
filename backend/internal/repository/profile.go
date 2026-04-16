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

const profileSelectCols = `id, user_id, full_name, phone, email,
    current_city, current_city_code, target_position,
    experience_years, COALESCE(preferred_company_types::text[], '{}'),
    created_at, updated_at`

func scanProfile(row pgx.Row, p *model.Profile) error {
	return row.Scan(
		&p.ID, &p.UserID, &p.FullName, &p.Phone, &p.Email,
		&p.CurrentCity, &p.CurrentCityCode, &p.TargetPosition,
		&p.ExperienceYears, &p.PreferredCompanyTypes,
		&p.CreatedAt, &p.UpdatedAt,
	)
}

func (r *ProfileRepo) Get(ctx context.Context, userID int64) (*model.Profile, error) {
	q := `SELECT ` + profileSelectCols + ` FROM user_profile WHERE user_id = $1`
	p := &model.Profile{}
	err := scanProfile(r.pool.QueryRow(ctx, q, userID), p)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ProfileRepo) Upsert(ctx context.Context, userID int64, in model.ProfileUpsertInput) (*model.Profile, error) {
	q := `
        INSERT INTO user_profile (
            user_id, full_name, phone, email,
            current_city, current_city_code, target_position,
            experience_years, preferred_company_types
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::company_type_enum[])
        ON CONFLICT (user_id) DO UPDATE SET
            full_name = EXCLUDED.full_name,
            phone = EXCLUDED.phone,
            email = EXCLUDED.email,
            current_city = EXCLUDED.current_city,
            current_city_code = EXCLUDED.current_city_code,
            target_position = EXCLUDED.target_position,
            experience_years = EXCLUDED.experience_years,
            preferred_company_types = EXCLUDED.preferred_company_types,
            updated_at = NOW()
        RETURNING ` + profileSelectCols
	types := in.PreferredCompanyTypes
	if types == nil {
		types = []string{}
	}
	p := &model.Profile{}
	err := scanProfile(r.pool.QueryRow(ctx, q,
		userID, in.FullName, in.Phone, in.Email,
		in.CurrentCity, in.CurrentCityCode, in.TargetPosition,
		in.ExperienceYears, types,
	), p)
	if err != nil {
		return nil, err
	}
	return p, nil
}
