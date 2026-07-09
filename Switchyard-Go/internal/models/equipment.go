package models

import (
	"time"

	"github.com/google/uuid"
)

type EquipmentType string
type EquipmentStatus string

const (
	EquipmentTypeTruck   EquipmentType = "truck"
	EquipmentTypeTractor EquipmentType = "tractor"
	EquipmentTypeTrailer EquipmentType = "trailer"

	EquipmentStatusAvailable   EquipmentStatus = "available"
	EquipmentStatusAssigned    EquipmentStatus = "assigned"
	EquipmentStatusMaintenance EquipmentStatus = "maintenance"
	EquipmentStatusBreakdown   EquipmentStatus = "breakdown"
)

type Equipment struct {
	ID              uuid.UUID       `json:"id"`
	UnitID          string          `json:"unit_id"`
	EquipmentType   EquipmentType   `json:"equipment_type"`
	HomeWarehouseID string          `json:"home_warehouse_id"`
	Status          EquipmentStatus `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
	// EmptyReturnUntil is set when equipment finishes a run with no deadhead pairing
	// secured: it stays Assigned (there is no location tracking to confirm arrival)
	// until this estimated return time passes, mirroring the driver's Empty Return
	// ETA. Any explicit UpdateStatus call clears it.
	EmptyReturnUntil *time.Time `json:"empty_return_until,omitempty"`
}

type MaintenanceRecord struct {
	ID              uuid.UUID  `json:"id"`
	EquipmentID     uuid.UUID  `json:"equipment_id"`
	Description     string     `json:"description"`
	ScheduledAt     time.Time  `json:"scheduled_at"`
	EstimatedReturn *time.Time `json:"estimated_return,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type BreakdownType string

const (
	BreakdownTypeDepot    BreakdownType = "depot"
	BreakdownTypeRoadside BreakdownType = "roadside"
)

type BreakdownRecord struct {
	ID            uuid.UUID     `json:"id"`
	EquipmentID   uuid.UUID     `json:"equipment_id"`
	BreakdownType BreakdownType `json:"breakdown_type"`
	LocationDesc  *string       `json:"location_desc,omitempty"`
	DriverID      *uuid.UUID    `json:"driver_id,omitempty"`
	LoadAttached  bool          `json:"load_attached"`
	ReportedAt    time.Time     `json:"reported_at"`
	ResolvedAt    *time.Time    `json:"resolved_at,omitempty"`
	TowCost       *float64      `json:"tow_cost,omitempty"`
}
