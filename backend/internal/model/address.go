package model

import "time"

type HomeAddress struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Alias     string    `json:"alias"`
	Address   string    `json:"address"`
	Province  *string   `json:"province,omitempty"`
	City      *string   `json:"city,omitempty"`
	District  *string   `json:"district,omitempty"`
	Longitude float64   `json:"longitude"`
	Latitude  float64   `json:"latitude"`
	IsDefault bool      `json:"is_default"`
	Note      *string   `json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HomeAddressCreateInput struct {
	Alias     string   `json:"alias" binding:"required,max=64"`
	Address   string   `json:"address" binding:"required,max=512"`
	Province  *string  `json:"province" binding:"omitempty,max=32"`
	City      *string  `json:"city" binding:"omitempty,max=32"`
	District  *string  `json:"district" binding:"omitempty,max=32"`
	Longitude float64  `json:"longitude" binding:"required,gte=-180,lte=180"`
	Latitude  float64  `json:"latitude" binding:"required,gte=-90,lte=90"`
	IsDefault bool     `json:"is_default"`
	Note      *string  `json:"note" binding:"omitempty"`
}

type HomeAddressUpdateInput struct {
	Alias     *string  `json:"alias" binding:"omitempty,max=64"`
	Address   *string  `json:"address" binding:"omitempty,max=512"`
	Province  *string  `json:"province" binding:"omitempty,max=32"`
	City      *string  `json:"city" binding:"omitempty,max=32"`
	District  *string  `json:"district" binding:"omitempty,max=32"`
	Longitude *float64 `json:"longitude" binding:"omitempty,gte=-180,lte=180"`
	Latitude  *float64 `json:"latitude" binding:"omitempty,gte=-90,lte=90"`
	IsDefault *bool    `json:"is_default"`
	Note      *string  `json:"note"`
}
