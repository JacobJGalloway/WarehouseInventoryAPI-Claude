package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/JacobJGalloway/switchyard-go/internal/models"
)

type AssignmentRepo struct{ db *pgxpool.Pool }

func NewAssignmentRepo(db *pgxpool.Pool) *AssignmentRepo { return &AssignmentRepo{db: db} }

const assignmentCols = `id, driver_id, plan_bol_id, equipment_id, base_rate_per_mile,
	assigned_at, departed_at, fulfilled_at, deadhead_confirmed_at,
	segment_start_stop_id, segment_end_stop_id, transfer_reason, notes, transferred_at,
	estimated_run_hours`

func scanAssignment(row interface{ Scan(...any) error }) (*models.DriverBOLAssignment, error) {
	a := &models.DriverBOLAssignment{}
	err := row.Scan(
		&a.ID, &a.DriverID, &a.PlanBOLID, &a.EquipmentID, &a.BaseRatePerMile,
		&a.AssignedAt, &a.DepartedAt, &a.FulfilledAt, &a.DeadheadConfirmedAt,
		&a.SegmentStartStopID, &a.SegmentEndStopID, &a.TransferReason, &a.Notes, &a.TransferredAt,
		&a.EstimatedRunHours,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *AssignmentRepo) Create(ctx context.Context, a *models.DriverBOLAssignment) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO driver_bol_assignment (`+assignmentCols+`)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		a.ID, a.DriverID, a.PlanBOLID, a.EquipmentID, a.BaseRatePerMile,
		a.AssignedAt, a.DepartedAt, a.FulfilledAt, a.DeadheadConfirmedAt,
		a.SegmentStartStopID, a.SegmentEndStopID, a.TransferReason, a.Notes, a.TransferredAt,
		a.EstimatedRunHours)
	return err
}

func (r *AssignmentRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.DriverBOLAssignment, error) {
	return scanAssignment(r.db.QueryRow(ctx,
		`SELECT `+assignmentCols+` FROM driver_bol_assignment WHERE id=$1`, id))
}

func (r *AssignmentRepo) GetByPlanBOL(ctx context.Context, planBOLID uuid.UUID) (*models.DriverBOLAssignment, error) {
	return scanAssignment(r.db.QueryRow(ctx,
		`SELECT `+assignmentCols+` FROM driver_bol_assignment
		 WHERE plan_bol_id=$1 AND transferred_at IS NULL
		 ORDER BY assigned_at DESC LIMIT 1`, planBOLID))
}

func (r *AssignmentRepo) GetAllActive(ctx context.Context) ([]*models.DriverBOLAssignment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+assignmentCols+` FROM driver_bol_assignment
		 WHERE deadhead_confirmed_at IS NULL ORDER BY assigned_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.DriverBOLAssignment
	for rows.Next() {
		a, err := scanAssignment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AssignmentRepo) GetActiveByDriver(ctx context.Context, driverID uuid.UUID) (*models.DriverBOLAssignment, error) {
	return scanAssignment(r.db.QueryRow(ctx,
		`SELECT `+assignmentCols+` FROM driver_bol_assignment
		 WHERE driver_id=$1 AND deadhead_confirmed_at IS NULL
		 ORDER BY assigned_at DESC LIMIT 1`, driverID))
}

func (r *AssignmentRepo) GetCustodyChain(ctx context.Context, planBOLID uuid.UUID) ([]*models.DriverBOLAssignment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+assignmentCols+` FROM driver_bol_assignment
		 WHERE plan_bol_id=$1 ORDER BY assigned_at ASC`, planBOLID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.DriverBOLAssignment
	for rows.Next() {
		a, err := scanAssignment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AssignmentRepo) MarkDeparted(ctx context.Context, id uuid.UUID, departedAt time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE driver_bol_assignment SET departed_at=$2 WHERE id=$1`, id, departedAt)
	return err
}

func (r *AssignmentRepo) MarkFulfilled(ctx context.Context, id uuid.UUID, fulfilledAt time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE driver_bol_assignment SET fulfilled_at=$2 WHERE id=$1`, id, fulfilledAt)
	return err
}

func (r *AssignmentRepo) ConfirmDeadhead(ctx context.Context, id uuid.UUID, confirmedAt time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE driver_bol_assignment SET deadhead_confirmed_at=$2 WHERE id=$1`, id, confirmedAt)
	return err
}

func (r *AssignmentRepo) InitiateTransfer(ctx context.Context, id uuid.UUID, transferredAt time.Time, segmentEndStopID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE driver_bol_assignment
		 SET transferred_at=$2, segment_end_stop_id=$3
		 WHERE id=$1`,
		id, transferredAt, segmentEndStopID)
	return err
}
