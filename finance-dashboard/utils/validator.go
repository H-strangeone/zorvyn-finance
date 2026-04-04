package utils

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// package-level maps — allocated once, not on every call
var validRoles = map[string]bool{
	"viewer":  true,
	"analyst": true,
	"admin":   true,
}

var validSorts = map[string]bool{
	"date_asc":    true,
	"date_desc":   true,
	"amount_asc":  true,
	"amount_desc": true,
}

var validRequestedRoles = map[string]bool{
	"analyst": true,
	"admin":   true,
}

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func IsValidRole(role string) bool {
	return validRoles[role]
}

func IsValidTransactionType(t string) bool {
	return t == "income" || t == "expense"
}

func IsValidSortOption(sort string) bool {
	return validSorts[sort]
}

func IsValidRoleRequestAction(action string) bool {
	return action == "approve" || action == "reject"
}

func IsValidRequestedRole(role string) bool {
	return validRequestedRoles[role]
}

// IsValidUUID validates UUID format before any store lookup
// Prevents garbage input from reaching the store layer
func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}