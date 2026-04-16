package model

import "time"

// User 简化的用户对象（返回给前端）
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name,omitempty"`
	Email     string    `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginInput 登录请求
type LoginInput struct {
	Email    string `json:"email" binding:"required,email,max=128"`
	Password string `json:"password" binding:"required,min=1,max=128"`
}

// LoginResponse 登录返回
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
}
