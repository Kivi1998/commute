package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/haojia/commute/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDuplicate = errors.New("duplicate")

type CompanyRepo struct {
	pool *pgxpool.Pool
}

func NewCompanyRepo(pool *pgxpool.Pool) *CompanyRepo {
	return &CompanyRepo{pool: pool}
}

const companyColumns = `id, user_id, name, address, province, city, district,
    longitude::float8, latitude::float8,
    category::text, industry,
    status::text, source::text,
    ai_reason, note, created_at, updated_at`

func scanCompany(row pgx.Row, c *model.Company) error {
	return row.Scan(
		&c.ID, &c.UserID, &c.Name, &c.Address,
		&c.Province, &c.City, &c.District,
		&c.Longitude, &c.Latitude,
		&c.Category, &c.Industry,
		&c.Status, &c.Source,
		&c.AIReason, &c.Note,
		&c.CreatedAt, &c.UpdatedAt,
	)
}

func (r *CompanyRepo) List(ctx context.Context, userID int64, q model.CompanyListQuery) (model.CompanyListResult, error) {
	conds := []string{"user_id = $1", "deleted_at IS NULL"}
	args := []any{userID}
	i := 2

	if q.Status != nil && *q.Status != "" {
		conds = append(conds, fmt.Sprintf("status = $%d::company_status_enum", i))
		args = append(args, *q.Status)
		i++
	}
	if q.Category != nil && *q.Category != "" {
		conds = append(conds, fmt.Sprintf("category = $%d::company_type_enum", i))
		args = append(args, *q.Category)
		i++
	}
	if q.Keyword != nil && strings.TrimSpace(*q.Keyword) != "" {
		conds = append(conds, fmt.Sprintf("(name ILIKE $%d OR address ILIKE $%d)", i, i))
		args = append(args, "%"+strings.TrimSpace(*q.Keyword)+"%")
		i++
	}
	where := strings.Join(conds, " AND ")

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM company WHERE %s", where)
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return model.CompanyListResult{}, err
	}

	offset := (q.Page - 1) * q.PageSize
	listSQL := fmt.Sprintf(`
        SELECT %s FROM company
        WHERE %s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d
    `, companyColumns, where, i, i+1)
	args = append(args, q.PageSize, offset)

	rows, err := r.pool.Query(ctx, listSQL, args...)
	if err != nil {
		return model.CompanyListResult{}, err
	}
	defer rows.Close()

	list := make([]model.Company, 0)
	for rows.Next() {
		var c model.Company
		if err := scanCompany(rows, &c); err != nil {
			return model.CompanyListResult{}, err
		}
		list = append(list, c)
	}

	totalPages := int(total) / q.PageSize
	if int(total)%q.PageSize > 0 {
		totalPages++
	}
	return model.CompanyListResult{
		List: list,
		Pagination: model.Pagination{
			Page: q.Page, PageSize: q.PageSize, Total: total, TotalPages: totalPages,
		},
	}, rows.Err()
}

func (r *CompanyRepo) Get(ctx context.Context, userID, id int64) (*model.Company, error) {
	q := fmt.Sprintf(`
        SELECT %s FROM company
        WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
    `, companyColumns)
	c := &model.Company{}
	err := scanCompany(r.pool.QueryRow(ctx, q, id, userID), c)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CompanyRepo) Create(ctx context.Context, userID int64, in model.CompanyCreateInput) (*model.Company, error) {
	status := "watching"
	if in.Status != nil {
		status = *in.Status
	}
	source := "manual"
	if in.Source != nil {
		source = *in.Source
	}
	q := fmt.Sprintf(`
        INSERT INTO company (
            user_id, name, address, province, city, district,
            longitude, latitude, category, industry, status, source, ai_reason, note
        )
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,
                $9::company_type_enum,$10,$11::company_status_enum,$12::company_source_enum,$13,$14)
        RETURNING %s
    `, companyColumns)

	c := &model.Company{}
	err := scanCompany(r.pool.QueryRow(ctx, q,
		userID, in.Name, in.Address, in.Province, in.City, in.District,
		in.Longitude, in.Latitude,
		in.Category, in.Industry, status, source,
		in.AIReason, in.Note,
	), c)
	if isUniqueViolation(err) {
		return nil, ErrDuplicate
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CompanyRepo) Update(ctx context.Context, userID, id int64, in model.CompanyUpdateInput) (*model.Company, error) {
	q := fmt.Sprintf(`
        UPDATE company SET
            name       = COALESCE($3, name),
            address    = COALESCE($4, address),
            province   = COALESCE($5, province),
            city       = COALESCE($6, city),
            district   = COALESCE($7, district),
            longitude  = COALESCE($8, longitude),
            latitude   = COALESCE($9, latitude),
            category   = COALESCE($10::company_type_enum, category),
            industry   = COALESCE($11, industry),
            status     = COALESCE($12::company_status_enum, status),
            note       = COALESCE($13, note),
            updated_at = NOW()
        WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
        RETURNING %s
    `, companyColumns)

	c := &model.Company{}
	err := scanCompany(r.pool.QueryRow(ctx, q,
		id, userID,
		in.Name, in.Address, in.Province, in.City, in.District,
		in.Longitude, in.Latitude,
		in.Category, in.Industry, in.Status, in.Note,
	), c)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CompanyRepo) UpdateStatus(ctx context.Context, userID, id int64, status string) (*model.Company, error) {
	q := fmt.Sprintf(`
        UPDATE company SET status = $3::company_status_enum, updated_at = NOW()
        WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
        RETURNING %s
    `, companyColumns)
	c := &model.Company{}
	err := scanCompany(r.pool.QueryRow(ctx, q, id, userID, status), c)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CompanyRepo) Delete(ctx context.Context, userID, id int64) error {
	cmd, err := r.pool.Exec(ctx, `
        UPDATE company SET deleted_at = NOW(), updated_at = NOW()
        WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
    `, id, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation
}
