package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/haojia/commute/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CommuteRepo struct {
	pool *pgxpool.Pool
}

func NewCommuteRepo(pool *pgxpool.Pool) *CommuteRepo {
	return &CommuteRepo{pool: pool}
}

// CreateQuery 保存一次查询会话，返回 query_id
func (r *CommuteRepo) CreateQuery(ctx context.Context, userID int64, in model.CommuteCalculateInput) (int64, error) {
	const q = `
        INSERT INTO commute_query (
            user_id, home_id, transport_modes,
            morning_strategy, morning_time, evening_strategy, evening_time,
            weekday, buffer_minutes
        )
        VALUES ($1, $2, $3::transport_mode_enum[],
                $4::time_strategy_enum, $5, $6::time_strategy_enum, $7,
                $8, $9)
        RETURNING id
    `
	var id int64
	err := r.pool.QueryRow(ctx, q,
		userID, in.HomeID, in.TransportModes,
		normalizeStrategy(in.Morning.Strategy), in.Morning.Time,
		normalizeStrategy(in.Evening.Strategy), in.Evening.Time,
		normalizeWeekday(in.Weekday), normalizeBuffer(in.BufferMinutes),
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func normalizeStrategy(s string) string {
	if s == "" {
		return "depart_at"
	}
	return s
}
func normalizeWeekday(w int) int {
	if w == 0 {
		return 1
	}
	return w
}
func normalizeBuffer(b int) int {
	if b == 0 {
		return 5
	}
	return b
}

// FindCachedResult 查找 7 天内未过期且未失败的缓存结果。
func (r *CommuteRepo) FindCachedResult(ctx context.Context, key CommuteResultKey) (*model.CommuteResult, error) {
	const q = `
        SELECT id, user_id, query_id, home_id, company_id,
               direction::text, transport_mode::text,
               depart_time::text, arrive_time::text,
               weekday, duration_min, duration_min_raw, distance_km::float8,
               cost_yuan::float8, transfer_count, polyline, route_detail,
               calculated_at, expires_at, is_failed, error_message
        FROM commute_result
        WHERE home_id = $1 AND company_id = $2
          AND transport_mode = $3::transport_mode_enum
          AND direction = $4::commute_direction_enum
          AND depart_time = $5::time
          AND weekday = $6
          AND expires_at > NOW()
          AND is_failed = FALSE
        ORDER BY calculated_at DESC LIMIT 1
    `
	row := r.pool.QueryRow(ctx, q,
		key.HomeID, key.CompanyID, key.TransportMode, key.Direction, key.DepartTime, key.Weekday,
	)
	var res model.CommuteResult
	err := scanCommuteResult(row, &res)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// InsertResult 写入一条计算结果（成功或失败都写入）。
func (r *CommuteRepo) InsertResult(ctx context.Context, rec CommuteResultInsert) (*model.CommuteResult, error) {
	const q = `
        INSERT INTO commute_result (
            user_id, query_id, home_id, company_id,
            direction, transport_mode, depart_time, arrive_time, weekday,
            duration_min, duration_min_raw, distance_km, cost_yuan, transfer_count,
            polyline, route_detail, expires_at, is_failed, error_message
        )
        VALUES ($1, $2, $3, $4,
                $5::commute_direction_enum, $6::transport_mode_enum, $7::time, $8::time, $9,
                $10, $11, $12, $13, $14,
                $15, $16, $17, $18, $19)
        RETURNING id, user_id, query_id, home_id, company_id,
                  direction::text, transport_mode::text,
                  depart_time::text, arrive_time::text,
                  weekday, duration_min, duration_min_raw, distance_km::float8,
                  cost_yuan::float8, transfer_count, polyline, route_detail,
                  calculated_at, expires_at, is_failed, error_message
    `
	var res model.CommuteResult
	routeDetail := rec.RouteDetail
	if routeDetail == nil {
		routeDetail = json.RawMessage("{}")
	}
	row := r.pool.QueryRow(ctx, q,
		rec.UserID, rec.QueryID, rec.HomeID, rec.CompanyID,
		rec.Direction, rec.TransportMode, rec.DepartTime, rec.ArriveTime, rec.Weekday,
		rec.DurationMin, rec.DurationMinRaw, rec.DistanceKM, rec.CostYuan, rec.TransferCount,
		rec.Polyline, routeDetail, rec.ExpiresAt, rec.IsFailed, rec.ErrorMessage,
	)
	if err := scanCommuteResult(row, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *CommuteRepo) GetResult(ctx context.Context, userID, id int64) (*model.CommuteResult, error) {
	const q = `
        SELECT id, user_id, query_id, home_id, company_id,
               direction::text, transport_mode::text,
               depart_time::text, arrive_time::text,
               weekday, duration_min, duration_min_raw, distance_km::float8,
               cost_yuan::float8, transfer_count, polyline, route_detail,
               calculated_at, expires_at, is_failed, error_message
        FROM commute_result WHERE id = $1 AND user_id = $2
    `
	var res model.CommuteResult
	err := scanCommuteResult(r.pool.QueryRow(ctx, q, id, userID), &res)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *CommuteRepo) ListResultsByQuery(ctx context.Context, userID, queryID int64) ([]model.CommuteResult, error) {
	const q = `
        SELECT id, user_id, query_id, home_id, company_id,
               direction::text, transport_mode::text,
               depart_time::text, arrive_time::text,
               weekday, duration_min, duration_min_raw, distance_km::float8,
               cost_yuan::float8, transfer_count, polyline, route_detail,
               calculated_at, expires_at, is_failed, error_message
        FROM commute_result
        WHERE user_id = $1 AND query_id = $2
        ORDER BY company_id, direction, transport_mode
    `
	rows, err := r.pool.Query(ctx, q, userID, queryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]model.CommuteResult, 0)
	for rows.Next() {
		var res model.CommuteResult
		if err := scanCommuteResult(rows, &res); err != nil {
			return nil, err
		}
		list = append(list, res)
	}
	return list, rows.Err()
}

type CommuteResultKey struct {
	HomeID        int64
	CompanyID     int64
	TransportMode string
	Direction     string
	DepartTime    string // HH:MM
	Weekday       int
}

type CommuteResultInsert struct {
	UserID         int64
	QueryID        *int64
	HomeID         int64
	CompanyID      int64
	Direction      string
	TransportMode  string
	DepartTime     string
	ArriveTime     string
	Weekday        int
	DurationMin    int
	DurationMinRaw int
	DistanceKM     float64
	CostYuan       *float64
	TransferCount  *int
	Polyline       string
	RouteDetail    json.RawMessage
	ExpiresAt      time.Time
	IsFailed       bool
	ErrorMessage   *string
}

func scanCommuteResult(row pgx.Row, r *model.CommuteResult) error {
	var queryID *int64
	if err := row.Scan(
		&r.ID, &r.UserID, &queryID, &r.HomeID, &r.CompanyID,
		&r.Direction, &r.TransportMode,
		&r.DepartTime, &r.ArriveTime,
		&r.Weekday, &r.DurationMin, &r.DurationMinRaw, &r.DistanceKM,
		&r.CostYuan, &r.TransferCount, &r.Polyline, &r.RouteDetail,
		&r.CalculatedAt, &r.ExpiresAt, &r.IsFailed, &r.ErrorMessage,
	); err != nil {
		return err
	}
	r.QueryID = queryID
	// depart_time / arrive_time 返回格式为 "HH:MM:SS"，截取前 5 位
	r.DepartTime = trimTime(r.DepartTime)
	r.ArriveTime = trimTime(r.ArriveTime)
	return nil
}

func trimTime(t string) string {
	if len(t) >= 5 {
		return t[:5]
	}
	return t
}

func (r *CommuteRepo) GetQuery(ctx context.Context, userID, id int64) (*model.CommuteQuery, error) {
	const q = `
        SELECT id, user_id, home_id,
               COALESCE(transport_modes::text[], '{}'),
               morning_strategy::text, morning_time::text,
               evening_strategy::text, evening_time::text,
               weekday, buffer_minutes, created_at
        FROM commute_query WHERE id = $1 AND user_id = $2
    `
	var cq model.CommuteQuery
	err := r.pool.QueryRow(ctx, q, id, userID).Scan(
		&cq.ID, &cq.UserID, &cq.HomeID, &cq.TransportModes,
		&cq.MorningStrategy, &cq.MorningTime,
		&cq.EveningStrategy, &cq.EveningTime,
		&cq.Weekday, &cq.BufferMinutes, &cq.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	cq.MorningTime = trimTime(cq.MorningTime)
	cq.EveningTime = trimTime(cq.EveningTime)
	return &cq, nil
}

// ListQueries 列出用户的所有查询会话（含聚合的公司名）
func (r *CommuteRepo) ListQueries(ctx context.Context, userID int64, limit int) ([]model.CommuteQueryListItem, error) {
	if limit <= 0 {
		limit = 50
	}
	const q = `
        SELECT
            q.id, q.user_id, q.home_id,
            COALESCE(q.transport_modes::text[], '{}'),
            q.morning_strategy::text, q.morning_time::text,
            q.evening_strategy::text, q.evening_time::text,
            q.weekday, q.buffer_minutes, q.created_at,
            COALESCE(h.alias, ''), COALESCE(h.address, ''),
            COALESCE((
                SELECT ARRAY_AGG(DISTINCT c.name ORDER BY c.name)
                FROM commute_result r
                JOIN company c ON c.id = r.company_id
                WHERE r.query_id = q.id
            ), '{}')
        FROM commute_query q
        LEFT JOIN home_address h ON h.id = q.home_id
        WHERE q.user_id = $1
        ORDER BY q.created_at DESC
        LIMIT $2
    `
	rows, err := r.pool.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.CommuteQueryListItem, 0)
	for rows.Next() {
		var it model.CommuteQueryListItem
		var companyNames []string
		err := rows.Scan(
			&it.ID, &it.UserID, &it.HomeID,
			&it.TransportModes,
			&it.MorningStrategy, &it.MorningTime,
			&it.EveningStrategy, &it.EveningTime,
			&it.Weekday, &it.BufferMinutes, &it.CreatedAt,
			&it.HomeAlias, &it.HomeAddress,
			&companyNames,
		)
		if err != nil {
			return nil, err
		}
		it.MorningTime = trimTime(it.MorningTime)
		it.EveningTime = trimTime(it.EveningTime)
		it.CompanyNames = companyNames
		it.CompanyCount = len(companyNames)
		list = append(list, it)
	}
	return list, rows.Err()
}

// DeleteQuery 删除查询会话（结果的 query_id 会 SET NULL）
func (r *CommuteRepo) DeleteQuery(ctx context.Context, userID, id int64) error {
	cmd, err := r.pool.Exec(ctx,
		`DELETE FROM commute_query WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// CleanupExpired 供定时任务调用
func (r *CommuteRepo) CleanupExpired(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
        DELETE FROM commute_result
        WHERE expires_at < NOW() - INTERVAL '7 days' AND is_failed = TRUE
    `)
	if err != nil {
		return fmt.Errorf("cleanup commute_result: %w", err)
	}
	return nil
}
