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
	RequestedRole Role          `json:"requestedRole"`
	Status        RequestStatus `json:"status"`
	Reason        string        `json:"reason"`
	ReviewedBy    string        `json:"reviewedBy"`
	ReviewNote    string        `json:"reviewNote"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}
