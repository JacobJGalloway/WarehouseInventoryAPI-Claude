package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/JacobJGalloway/switchyard-go/internal/models"
)

type DriverRepository interface {
	Create(ctx context.Context, d *models.Driver) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Driver, error)
	GetAll(ctx context.Context) ([]*models.Driver, error)
	Update(ctx context.Context, d *models.Driver) error
}

type HOSRepository interface {
	CreateLimit(ctx context.Context, l *models.HOSLimit) error
	GetLimitByStateAndCycle(ctx context.Context, stateCode string, cycleLabel string) (*models.HOSLimit, error)
	CreateWindow(ctx context.Context, w *models.HOSWindow) error
	GetWindowByDriver(ctx context.Context, driverID uuid.UUID) (*models.HOSWindow, error)
	UpdateWindow(ctx context.Context, w *models.HOSWindow) error
}

type EquipmentRepository interface {
	Create(ctx context.Context, e *models.Equipment) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Equipment, error)
	GetAll(ctx context.Context) ([]*models.Equipment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.EquipmentStatus) error
	// SetEmptyReturnUntil marks equipment as still Assigned but returning empty,
	// estimated available again at `until`. Status is left unchanged.
	SetEmptyReturnUntil(ctx context.Context, id uuid.UUID, until time.Time) error
	CreateMaintenanceRecord(ctx context.Context, r *models.MaintenanceRecord) error
	GetActiveMaintenanceByEquipment(ctx context.Context, equipmentID uuid.UUID) (*models.MaintenanceRecord, error)
	ResolveMaintenanceRecord(ctx context.Context, id uuid.UUID, completedAt time.Time) error
	CreateBreakdownRecord(ctx context.Context, r *models.BreakdownRecord) error
	GetActiveBreakdownByEquipment(ctx context.Context, equipmentID uuid.UUID) (*models.BreakdownRecord, error)
	ResolveBreakdownRecord(ctx context.Context, id uuid.UUID, resolvedAt time.Time, towCost *float64) error
}

type PlanBOLRepository interface {
	Create(ctx context.Context, p *models.PlanBOLRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.PlanBOLRecord, error)
	GetByStatus(ctx context.Context, status models.PlanBOLStatus) ([]*models.PlanBOLRecord, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.PlanBOLStatus) error
	SetSubmittedTransactionID(ctx context.Context, id uuid.UUID, txID string) error
	CreateStop(ctx context.Context, s *models.PlanBOLStop) error
	GetStopByID(ctx context.Context, stopID uuid.UUID) (*models.PlanBOLStop, error)
	GetStops(ctx context.Context, planBOLID uuid.UUID) ([]*models.PlanBOLStop, error)
	MarkStopProcessed(ctx context.Context, stopID uuid.UUID, processedAt time.Time, milesLeg *float64) error
	SetMilesDriven(ctx context.Context, id uuid.UUID, miles float64) error
	CreateSnapshot(ctx context.Context, s *models.TruckInventorySnapshot) error
	GetSnapshots(ctx context.Context, planBOLID uuid.UUID) ([]*models.TruckInventorySnapshot, error)
	GetStatusHistory(ctx context.Context, planBOLID uuid.UUID) ([]*models.BOLStatusHistory, error)
}

type AssignmentRepository interface {
	Create(ctx context.Context, a *models.DriverBOLAssignment) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.DriverBOLAssignment, error)
	// GetByPlanBOL returns the current active assignment for a BOL (transferred_at IS NULL).
	GetByPlanBOL(ctx context.Context, planBOLID uuid.UUID) (*models.DriverBOLAssignment, error)
	// GetAllActive returns all assignments where deadhead_confirmed_at IS NULL.
	// This covers Pending Dispatch, In Transit, Delivered, and Transfer Deadhead board states.
	GetAllActive(ctx context.Context) ([]*models.DriverBOLAssignment, error)
	// GetActiveByDriver returns the most recent assignment for a driver where
	// deadhead_confirmed_at IS NULL. Returns nil, nil when the driver has no active run.
	GetActiveByDriver(ctx context.Context, driverID uuid.UUID) (*models.DriverBOLAssignment, error)
	// GetCustodyChain returns all assignments for a BOL ordered by assigned_at ascending.
	// Single-driver BOLs return one record; transferred BOLs return the full chain.
	GetCustodyChain(ctx context.Context, planBOLID uuid.UUID) ([]*models.DriverBOLAssignment, error)
	MarkDeparted(ctx context.Context, id uuid.UUID, departedAt time.Time) error
	MarkFulfilled(ctx context.Context, id uuid.UUID, fulfilledAt time.Time) error
	ConfirmDeadhead(ctx context.Context, id uuid.UUID, confirmedAt time.Time) error
	// InitiateTransfer closes the outgoing driver's segment by stamping transferred_at
	// and setting segment_end_stop_id to the transfer stop.
	InitiateTransfer(ctx context.Context, id uuid.UUID, transferredAt time.Time, segmentEndStopID uuid.UUID) error
}

type PairingRepository interface {
	Create(ctx context.Context, p *models.PlanBOLPairing) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.PlanBOLPairing, error)
	GetByActiveBOL(ctx context.Context, activeBOLID uuid.UUID) (*models.PlanBOLPairing, error)
	// GetEligible returns pairings whose origin warehouse is nearest to location
	// and whose earliest_valid_at falls within the given estimated completion window.
	GetEligible(ctx context.Context, location string, estimatedCompletion time.Time) ([]*models.PlanBOLPairing, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.PairingStatus) error
}

type InvoiceRepository interface {
	CreateConfirmation(ctx context.Context, c *models.DeliveryConfirmation) error
	GetConfirmation(ctx context.Context, id uuid.UUID) (*models.DeliveryConfirmation, error)
	GetConfirmationsByBOL(ctx context.Context, planBOLID uuid.UUID) ([]*models.DeliveryConfirmation, error)
	SetConfirmationInvoice(ctx context.Context, confirmationID uuid.UUID, invoiceID uuid.UUID) error
	CreateInvoice(ctx context.Context, i *models.InternalInvoice) error
	GetInvoice(ctx context.Context, id uuid.UUID) (*models.InternalInvoice, error)
	GetInvoicesByStore(ctx context.Context, storeID string) ([]*models.InternalInvoice, error)
}

type WarehouseRepository interface {
	GetAll(ctx context.Context) ([]*models.Warehouse, error)
	GetByRegion(ctx context.Context, region string) ([]*models.Warehouse, error)
	Create(ctx context.Context, w *models.Warehouse) error
}

type AnalyticsQuerier interface {
	BOLsByStatus(ctx context.Context) ([]models.BOLStatusCount, error)
	StopCompletionRate(ctx context.Context) (float64, error)
	FulfilledInWindow(ctx context.Context, since time.Time) (int, error)
	OperatingCostByBOL(ctx context.Context) ([]models.BOLOperatingCost, error)
	OperatingCostByDriver(ctx context.Context) ([]models.DriverOperatingCost, error)
	OperatingCostByWarehouse(ctx context.Context) ([]models.WarehouseOperatingCost, error)
}
