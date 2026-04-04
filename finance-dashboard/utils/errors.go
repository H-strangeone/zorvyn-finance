package utils

import "fmt"

type AppError struct {
	Code       string // machine-readable: "VALIDATION_ERROR"
	Message    string // short, user-facing: "Validation failed"
	Details    string // deeper explanation: "amount must be greater than 0"
	StatusCode int
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
}

func NewValidationError(details string) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "Validation failed",
		Details:    details,
		StatusCode: 400,
	}
}

func NewUnauthorizedError(details string) *AppError {
	return &AppError{
		Code:       "UNAUTHORIZED",
		Message:    "Authentication required",
		Details:    details,
		StatusCode: 401,
	}
}

func NewForbiddenError(details string) *AppError {
	return &AppError{
		Code:       "FORBIDDEN",
		Message:    "Access denied",
		Details:    details,
		StatusCode: 403,
	}
}

func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       "NOT_FOUND",
		Message:    "Resource not found",
		Details:    fmt.Sprintf("%s not found", resource),
		StatusCode: 404,
	}
}

func NewConflictError(details string) *AppError {
	return &AppError{
		Code:       "CONFLICT",
		Message:    "Request conflicts with current state",
		Details:    details,
		StatusCode: 409,
	}
}

func NewInternalError(details string) *AppError {
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "An unexpected error occurred",
		Details:    details,
		StatusCode: 500,
	}
}