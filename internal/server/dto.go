package server

import "github.com/ashtishad/xm/internal/domain"

// ErrorResponse represents a standardized error message structure.
// @Description ErrorResponse provides a consistent error format.
type ErrorResponse struct {
	Error string `json:"error"`
}

// RegisterUserRequest holds the data for creating a new user.
// @Description RegisterUserRequest validates input for user registration.
// @Description Name must be 5-100 characters long.
// @Description Email must be a valid email address.
// @Description Password must be at least 8 characters long.
type RegisterUserRequest struct {
	Name     string `json:"name" form:"name" binding:"required,min=5,max=100"`
	Email    string `json:"email" form:"email" binding:"required,email"`
	Password string `json:"password" form:"password" binding:"required,min=8,max=100"`
}

// RegisterUserResponse contains the user data returned after successful registration.
// @Description RegisterUserResponse includes the created user's details.
type RegisterUserResponse struct {
	User domain.User `json:"user"`
}

// LoginRequest represents the credentials required for user authentication.
// @Description LoginRequest validates input for user login.
// @Description Email must be a valid email address.
// @Description Password is required.
type LoginRequest struct {
	Email    string `form:"email" json:"email" binding:"required,email"`
	Password string `form:"password" json:"password" binding:"required"`
}

// LoginResponse contains the user data returned after successful login.
// @Description LoginResponse includes the authenticated user's details.
type LoginResponse struct {
	User domain.User `json:"user"`
}

// CreateCompanyRequest holds the data for creating a new company
type CreateCompanyRequest struct {
	Name              string  `json:"name" binding:"required,max=15"`
	Description       *string `json:"description" binding:"omitempty,max=3000"`
	AmountOfEmployees int     `json:"amountOfEmployees" binding:"required,min=1"`
	Registered        bool    `json:"registered" binding:"required"`
	Type              string  `json:"type" binding:"required,oneof=Corporations NonProfit Cooperative 'Sole Proprietorship'"`
}

// UpdateCompanyRequest holds the data for updating a company
type UpdateCompanyRequest struct {
	Name              *string `json:"name" binding:"omitempty,max=15"`
	Description       *string `json:"description" binding:"omitempty,max=3000"`
	AmountOfEmployees *int    `json:"amountOfEmployees" binding:"omitempty,min=1"`
	Registered        *bool   `json:"registered" binding:"omitempty"`
	Type              *string `json:"type" binding:"omitempty,oneof=Corporations NonProfit Cooperative 'Sole Proprietorship'"`
}
