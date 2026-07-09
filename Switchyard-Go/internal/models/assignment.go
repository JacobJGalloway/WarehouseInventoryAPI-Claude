package models

import (
	"time"

	"github.com/google/uuid"
)

type TransferReason string

const (
	TransferReasonHOSLimit   TransferReason = "hos_limit"
	TransferReasonEmergency  TransferReason = "emergency"
	TransferReasonPlanned    TransferReason = "planned"
	TransferReasonOther      TransferReason = "other"
)

// DriverBOLAssignment links a driver, a planned BOL, and a piece of equipment
// for a single run segment. A BOL with no transfer has one record (origin → destination).
// A BOL with a mid-route transfer has two: origin → transfer stop, transfer stop → destination.
type DriverBOLAssignment struct {
	ID                  uuid.UUID      `json:"id"`
	DriverID            uuid.UUID      `json:"driver_id"`
	PlanBOLID           uuid.UUID      `json:"plan_bol_id"`
	EquipmentID         uuid.UUID      `json:"equipment_id"`
	BaseRatePerMile     float64        `json:"base_rate_per_mile"`
	AssignedAt          time.Time      `json:"assigned_at"`
	DepartedAt          *time.Time     `json:"departed_at,omitempty"`
	FulfilledAt         *time.Time     `json:"fulfilled_at,omitempty"`
	DeadheadConfirmedAt *time.Time     `json:"deadhead_confirmed_at,omitempty"`
	SegmentStartStopID  *uuid.UUID     `json:"segment_start_stop_id,omitempty"`
	SegmentEndStopID    *uuid.UUID     `json:"segment_end_stop_id,omitempty"`
	TransferReason      *TransferReason `json:"transfer_reason,omitempty"`
	Notes               *string        `json:"notes,omitempty"`
	TransferredAt       *time.Time     `json:"transferred_at,omitempty"`
	// EstimatedRunHours is the dispatcher-supplied run-length estimate captured at
	// assignment time (same value used for the HOS eligibility check). Combined with
	// DepartedAt it lets the board derive an estimated fulfillment time for the
	// deadhead cutoff-window check, without any live location tracking.
	EstimatedRunHours   *float64       `json:"estimated_run_hours,omitempty"`
}
