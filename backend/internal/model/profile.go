package model

import "time"

type Profile struct {
	ID                    int64     `json:"id"`
	UserID                int64     `json:"user_id"`
	CurrentCity           string    `json:"current_city"`
	CurrentCityCode       *string   `json:"current_city_code,omitempty"`
	TargetPosition        string    `json:"target_position"`
	ExperienceYears       *int16    `json:"experience_years,omitempty"`
	PreferredCompanyTypes []string  `json:"preferred_company_types"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type ProfileUpsertInput struct {
	CurrentCity           string   `json:"current_city" binding:"required,max=64"`
	CurrentCityCode       *string  `json:"current_city_code" binding:"omitempty,max=16"`
	TargetPosition        string   `json:"target_position" binding:"required,max=128"`
	ExperienceYears       *int16   `json:"experience_years" binding:"omitempty,min=0,max=30"`
	PreferredCompanyTypes []string `json:"preferred_company_types" binding:"omitempty,dive,oneof=big_tech mid_tech startup foreign other"`
}
