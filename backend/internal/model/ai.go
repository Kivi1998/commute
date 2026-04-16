package model

import "time"

// AIRecommendInput 推荐请求
type AIRecommendInput struct {
	City            string   `json:"city" binding:"required,max=64"`
	Position        string   `json:"position" binding:"required,max=128"`
	ExperienceYears *int     `json:"experience_years" binding:"omitempty,min=0,max=30"`
	CompanyTypes    []string `json:"company_types" binding:"omitempty,dive,oneof=big_tech mid_tech startup foreign other"`
	Count           int      `json:"count" binding:"omitempty,min=5,max=50"`
	ForceRefresh    bool     `json:"force_refresh"`
	// ExcludeNames 要排除的公司名（用户已关注的或上轮已见到的）。传入时会自动 force_refresh 绕开缓存。
	ExcludeNames []string `json:"exclude_names" binding:"omitempty,max=100,dive,max=128"`
}

// AIRecommendedCompany 单条推荐（供前端展示）
type AIRecommendedCompany struct {
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Industry    string  `json:"industry"`
	AddressHint string  `json:"address_hint"`
	Reason      string  `json:"reason"`

	// 地址二次校验后的精确坐标（可能为空表示未匹配到 POI）
	ResolvedAddress   *string  `json:"resolved_address,omitempty"`
	ResolvedLongitude *float64 `json:"resolved_longitude,omitempty"`
	ResolvedLatitude  *float64 `json:"resolved_latitude,omitempty"`
	ResolvedProvince  *string  `json:"resolved_province,omitempty"`
	ResolvedCity      *string  `json:"resolved_city,omitempty"`
	ResolvedDistrict  *string  `json:"resolved_district,omitempty"`
	LocationConfident bool     `json:"location_confident"` // true 表示坐标可用
}

// AIRecommendResult 推荐结果（聚合）
type AIRecommendResult struct {
	FromCache bool                   `json:"from_cache"`
	CachedAt  *time.Time             `json:"cached_at,omitempty"`
	ExpiresAt *time.Time             `json:"expires_at,omitempty"`
	Companies []AIRecommendedCompany `json:"companies"`
	TokenInput  int                  `json:"token_input,omitempty"`
	TokenOutput int                  `json:"token_output,omitempty"`
}
