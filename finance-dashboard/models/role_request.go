package models

import "time"


type RequestStatus string

const (
	StatusPending  RequestStatus = "pending"
	StatusApproved RequestStatus = "approved"
	StatusRejected RequestStatus = "rejected"
)

type RoleRequest struct {
	ID            string        `json:"id"`
	UserID        string        `json:"userId"`
	RequestedRole Role          `json:"requestedRole"` // analyst or admin only
	Status        RequestStatus `json:"status"`
	Reason        string        `json:"reason"`        // optional, from user
	ReviewedBy    string        `json:"reviewedBy"`    // admin userId, empty until reviewed
	ReviewNote    string        `json:"reviewNote"`    // optional, from admin
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}