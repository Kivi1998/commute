package model

import (
	"encoding/json"
	"time"
)

// --- 查询输入 ---

type CommuteTimeSpec struct {
	Strategy string `json:"strategy" binding:"omitempty,oneof=depart_at arrive_by"` // MVP 只实现 depart_at
	Time     string `json:"time" binding:"required"`                               // HH:MM
}

type CommuteCalculateInput struct {
	HomeID         int64           `json:"home_id" binding:"required,gt=0"`
	CompanyIDs     []int64         `json:"company_ids" binding:"required,min=1,max=50,dive,gt=0"`
	TransportModes []string        `json:"transport_modes" binding:"required,min=1,dive,oneof=transit driving cycling walking"`
	Morning        CommuteTimeSpec `json:"morning" binding:"required"`
	Evening        CommuteTimeSpec `json:"evening" binding:"required"`
	Weekday        int             `json:"weekday" binding:"omitempty,min=1,max=7"`
	BufferMinutes  int             `json:"buffer_minutes" binding:"omitempty,min=0,max=60"`
	ForceRefresh   bool            `json:"force_refresh"`
	SaveQuery      bool            `json:"save_query"`
}

// --- 查询记录 ---

type CommuteQuery struct {
	ID               int64     `json:"id"`
	UserID           int64     `json:"user_id"`
	HomeID           int64     `json:"home_id"`
	TransportModes   []string  `json:"transport_modes"`
	MorningStrategy  string    `json:"morning_strategy"`
	MorningTime      string    `json:"morning_time"`
	EveningStrategy  string    `json:"evening_strategy"`
	EveningTime      string    `json:"evening_time"`
	Weekday          int       `json:"weekday"`
	BufferMinutes    int       `json:"buffer_minutes"`
	CreatedAt        time.Time `json:"created_at"`
}

// CommuteQueryListItem 列表项（带聚合统计）
type CommuteQueryListItem struct {
	CommuteQuery
	HomeAlias    string   `json:"home_alias"`
	HomeAddress  string   `json:"home_address"`
	CompanyCount int      `json:"company_count"`
	CompanyNames []string `json:"company_names"`
}

// --- 计算结果 ---

type CommuteResult struct {
	ID              int64           `json:"id"`
	UserID          int64           `json:"user_id"`
	QueryID         *int64          `json:"query_id,omitempty"`
	HomeID          int64           `json:"home_id"`
	CompanyID       int64           `json:"company_id"`
	Direction       string          `json:"direction"`
	TransportMode   string          `json:"transport_mode"`
	DepartTime      string          `json:"depart_time"`
	ArriveTime      string          `json:"arrive_time"`
	Weekday         int             `json:"weekday"`
	DurationMin     int             `json:"duration_min"`
	DurationMinRaw  int             `json:"duration_min_raw"`
	DistanceKM      float64         `json:"distance_km"`
	CostYuan        *float64        `json:"cost_yuan,omitempty"`
	TransferCount   *int            `json:"transfer_count,omitempty"`
	RouteDetail     json.RawMessage `json:"route_detail"`
	CalculatedAt    time.Time       `json:"calculated_at"`
	ExpiresAt       time.Time       `json:"expires_at"`
	IsFailed        bool            `json:"is_failed"`
	ErrorMessage    *string         `json:"error_message,omitempty"`
	FromCache       bool            `json:"from_cache"`
}

// --- 响应聚合 ---

type CommuteCalculateResponse struct {
	QueryID *int64               `json:"query_id,omitempty"`
	Home    *HomeAddress         `json:"home"`
	Weekday int                  `json:"weekday"`
	BufferMinutes int            `json:"buffer_minutes"`
	Results []CompanyCommute     `json:"results"`
	Summary CommuteSummary       `json:"summary"`
}

type CompanyCommute struct {
	CompanyID        int64                  `json:"company_id"`
	CompanyName      string                 `json:"company_name"`
	CompanyLongitude float64                `json:"company_longitude"`
	CompanyLatitude  float64                `json:"company_latitude"`
	Items            []CommuteResultItem    `json:"items"`
	Errors           []CommuteCalcError     `json:"errors"`
}

type CommuteResultItem struct {
	Direction      string   `json:"direction"`
	TransportMode  string   `json:"transport_mode"`
	DepartTime     string   `json:"depart_time"`
	ArriveTime     string   `json:"arrive_time"`
	DurationMin    int      `json:"duration_min"`
	DurationMinRaw int      `json:"duration_min_raw"`
	DistanceKM     float64  `json:"distance_km"`
	CostYuan       *float64 `json:"cost_yuan,omitempty"`
	TransferCount  *int     `json:"transfer_count,omitempty"`
	FromCache      bool     `json:"from_cache"`
	ResultID       int64    `json:"result_id"`
}

type CommuteCalcError struct {
	Direction     string `json:"direction"`
	TransportMode string `json:"transport_mode"`
	Message       string `json:"message"`
}

type CommuteSummary struct {
	TotalCompanies    int `json:"total_companies"`
	TotalCalculations int `json:"total_calculations"`
	CacheHits         int `json:"cache_hits"`
	Failures          int `json:"failures"`
}
