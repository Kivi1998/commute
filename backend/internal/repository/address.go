package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/haojia/commute/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AddressRepo struct {
	pool *pgxpool.Pool
}

func NewAddressRepo(pool *pgxpool.Pool) *AddressRepo {
	return &AddressRepo{pool: pool}
}

const addressColumns = `id, user_id, alias, address, province, city, district,
        longitude::float8, latitude::float8, is_default, note, created_at, updated_at`

func scanAddress(row pgx.Row, a *model.HomeAddress) error {
	return row.Scan(
		&a.ID, &a.UserID, &a.Alias, &a.Address,
		&a.Province, &a.City, &a.District,
		&a.Longitude, &a.Latitude,
		&a.IsDefault, &a.Note,
		&a.CreatedAt, &a.UpdatedAt,
	)
}

func (r *AddressRepo) List(ctx context.Context, userID int64) ([]model.HomeAddress, error) {
	q := fmt.Sprintf(`
        SELECT %s FROM home_address
        WHERE user_id = $1 AND deleted_at IS NULL
        ORDER BY is_default DESC, created_at ASC
    `, addressColumns)
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.HomeAddress, 0)
	for rows.Next() {
		var a model.HomeAddress
		if err := scanAddress(rows, &a); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *AddressRepo) Get(ctx context.Context, userID, id int64) (*model.HomeAddress, error) {
	q := fmt.Sprintf(`
        SELECT %s FROM home_address
        WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
    `, addressColumns)
	a := &model.HomeAddress{}
	err := scanAddress(r.pool.QueryRow(ctx, q, id, userID), a)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *AddressRepo) Create(ctx context.Context, userID int64, in model.HomeAddressCreateInput) (*model.HomeAddress, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if in.IsDefault {
		if err := clearDefault(ctx, tx, userID); err != nil {
			return nil, err
		}
	}

	q := fmt.Sprintf(`
        INSERT INTO home_address (
            user_id, alias, address, province, city, district,
            longitude, latitude, is_default, note
        )
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
        RETURNING %s
    `, addressColumns)

	a := &model.HomeAddress{}
	err = scanAddress(tx.QueryRow(ctx, q,
		userID, in.Alias, in.Address, in.Province, in.City, in.District,
		in.Longitude, in.Latitude, in.IsDefault, in.Note,
	), a)
	if err != nil {
		return nil, err
	}
	return a, tx.Commit(ctx)
}

func (r *AddressRepo) Update(ctx context.Context, userID, id int64, in model.HomeAddressUpdateInput) (*model.HomeAddress, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if in.IsDefault != nil && *in.IsDefault {
		if err := clearDefault(ctx, tx, userID); err != nil {
			return nil, err
		}
	}

	q := fmt.Sprintf(`
        UPDATE home_address SET
            alias       = COALESCE($3, alias),
            address     = COALESCE($4, address),
            province    = COALESCE($5, province),
            city        = COALESCE($6, city),
            district    = COALESCE($7, district),
            longitude   = COALESCE($8, longitude),
            latitude    = COALESCE($9, latitude),
            is_default  = COALESCE($10, is_default),
            note        = COALESCE($11, note),
            updated_at  = NOW()
        WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
        RETURNING %s
    `, addressColumns)

	a := &model.HomeAddress{}
	err = scanAddress(tx.QueryRow(ctx, q,
		id, userID, in.Alias, in.Address, in.Province, in.City, in.District,
		in.Longitude, in.Latitude, in.IsDefault, in.Note,
	), a)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return a, tx.Commit(ctx)
}

func (r *AddressRepo) SetDefault(ctx context.Context, userID, id int64) (*model.HomeAddress, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := clearDefault(ctx, tx, userID); err != nil {
		return nil, err
	}

	q := fmt.Sprintf(`
        UPDATE home_address SET is_default = TRUE, updated_at = NOW()
        WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
        RETURNING %s
    `, addressColumns)

	a := &model.HomeAddress{}
	err = scanAddress(tx.QueryRow(ctx, q, id, userID), a)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return a, tx.Commit(ctx)
}

func (r *AddressRepo) Delete(ctx context.Context, userID, id int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var wasDefault bool
	err = tx.QueryRow(ctx, `
        UPDATE home_address SET deleted_at = NOW(), is_default = FALSE, updated_at = NOW()
        WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
        RETURNING is_default OR TRUE -- 总是返回一行以区分不存在
    `, id, userID).Scan(&wasDefault)
	_ = wasDefault
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	// 若删除的是原默认地址且还有剩余地址，把最早创建的提升为默认
	_, err = tx.Exec(ctx, `
        UPDATE home_address SET is_default = TRUE, updated_at = NOW()
        WHERE id = (
            SELECT id FROM home_address
            WHERE user_id = $1 AND deleted_at IS NULL
            ORDER BY created_at ASC LIMIT 1
        )
        AND NOT EXISTS (
            SELECT 1 FROM home_address
            WHERE user_id = $1 AND deleted_at IS NULL AND is_default = TRUE
        )
    `, userID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func clearDefault(ctx context.Context, tx pgx.Tx, userID int64) error {
	_, err := tx.Exec(ctx, `
        UPDATE home_address SET is_default = FALSE, updated_at = NOW()
        WHERE user_id = $1 AND is_default = TRUE AND deleted_at IS NULL
    `, userID)
	return err
}
