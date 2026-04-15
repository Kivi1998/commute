package model

import "time"

type Company struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Province  *string   `json:"province,omitempty"`
	City      *string   `json:"city,omitempty"`
	District  *string   `json:"district,omitempty"`
	Longitude float64   `json:"longitude"`
	Latitude  float64   `json:"latitude"`
	Category  *string   `json:"category,omitempty"`
	Industry  *string   `json:"industry,omitempty"`
	Status    string    `json:"status"`
	Source    string    `json:"source"`
	AIReason  *string   `json:"ai_reason,omitempty"`
	Note      *string   `json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CompanyCreateInput struct {
	Name      string   `json:"name" binding:"required,max=128"`
	Address   string   `json:"address" binding:"required,max=512"`
	Province  *string  `json:"province" binding:"omitempty,max=32"`
	City      *string  `json:"city" binding:"omitempty,max=32"`
	District  *string  `json:"district" binding:"omitempty,max=32"`
	Longitude float64  `json:"longitude" binding:"required,gte=-180,lte=180"`
	Latitude  float64  `json:"latitude" binding:"required,gte=-90,lte=90"`
	Category  *string  `json:"category" binding:"omitempty,oneof=big_tech mid_tech startup foreign other"`
	Industry  *string  `json:"industry" binding:"omitempty,max=64"`
	Status    *string  `json:"status" binding:"omitempty,oneof=watching applied interviewing offered rejected archived"`
	Source    *string  `json:"source" binding:"omitempty,oneof=ai_recommend manual"`
	AIReason  *string  `json:"ai_reason"`
	Note      *string  `json:"note"`
}

type CompanyUpdateInput struct {
	Name      *string  `json:"name" binding:"omitempty,max=128"`
	Address   *string  `json:"address" binding:"omitempty,max=512"`
	Province  *string  `json:"province" binding:"omitempty,max=32"`
	City      *string  `json:"city" binding:"omitempty,max=32"`
	District  *string  `json:"district" binding:"omitempty,max=32"`
	Longitude *float64 `json:"longitude" binding:"omitempty,gte=-180,lte=180"`
	Latitude  *float64 `json:"latitude" binding:"omitempty,gte=-90,lte=90"`
	Category  *string  `json:"category" binding:"omitempty,oneof=big_tech mid_tech startup foreign other"`
	Industry  *string  `json:"industry" binding:"omitempty,max=64"`
	Status    *string  `json:"status" binding:"omitempty,oneof=watching applied interviewing offered rejected archived"`
	Note      *string  `json:"note"`
}

type CompanyStatusInput struct {
	Status string `json:"status" binding:"required,oneof=watching applied interviewing offered rejected archived"`
}

type CompanyListQuery struct {
	Status   *string `form:"status"`
	Category *string `form:"category"`
	Keyword  *string `form:"keyword"`
	Page     int     `form:"page,default=1" binding:"gte=1"`
	PageSize int     `form:"page_size,default=50" binding:"gte=1,lte=100"`
}

type CompanyListResult struct {
	List       []Company  `json:"list"`
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type CompanyBatchInput struct {
	Companies []CompanyCreateInput `json:"companies" binding:"required,min=1,max=50,dive"`
}

type CompanyBatchResult struct {
	Created []Company        `json:"created"`
	Skipped []SkippedCompany `json:"skipped"`
	Warning *string          `json:"warning,omitempty"`
}

type SkippedCompany struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}
