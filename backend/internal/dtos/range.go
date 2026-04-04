package dtos

import "github.com/google/uuid"

// V1AllocateRangeRequest is the request to allocate a new range of IDs
type V1AllocateRangeRequest struct {
	OwnerID string `json:"owner_id" binding:"required"`
}

// V1AllocateRangeResponse is the response containing allocated ID range
type V1AllocateRangeResponse struct {
	ID            uuid.UUID `json:"id"`
	Start         uint64    `json:"start"`
	Last          uint64    `json:"last"`
	CurrentOffset uint64    `json:"current_offset"`
}

// V1UpdateRangeOffsetRequest is the request to update offset of an existing range
type V1UpdateRangeOffsetRequest struct {
	OwnerID string `json:"owner_id" binding:"required"`
}

// V1UpdateRangeOffsetResponse is the response after updating range offset
type V1UpdateRangeOffsetResponse struct {
	ID            uuid.UUID `json:"id"`
	CurrentOffset uint64    `json:"current_offset"`
}
