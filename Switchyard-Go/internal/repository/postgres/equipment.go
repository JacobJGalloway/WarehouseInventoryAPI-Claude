package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/JacobJGalloway/switchyard-go/internal/models"
)

type EquipmentRepo struct{ db *pgxpool.Pool }

func NewEquipmentRepo(db *pgxpool.Pool) *EquipmentRepo { return &EquipmentRepo{db: db} }

func (r *EquipmentRepo) Create(ctx context.Context, e *models.Equipment) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO equipment (id, unit_id, equipment_type, home_warehouse_id, status, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		e.ID, e.UnitID, string(e.EquipmentType), e.HomeWarehouseID, string(e.Status), e.CreatedAt)
	return err
}

func (r *EquipmentRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Equipment, error) {
	e := &models.Equipment{}
	var equipType, status string
	err := r.db.QueryRow(ctx,
		`SELECT id, unit_id, equipment_type, home_warehouse_id, status, created_at, empty_return_until
		 FROM equipment WHERE id=$1`, id).
		Scan(&e.ID, &e.UnitID, &equipType, &e.HomeWarehouseID, &status, &e.CreatedAt, &e.EmptyReturnUntil)
	if err != nil {
		return nil, err
	}
	e.EquipmentType = models.EquipmentType(equipType)
	e.Status = models.EquipmentStatus(status)
	return e, nil
}

func (r *EquipmentRepo) GetAll(ctx context.Context) ([]*models.Equipment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, unit_id, equipment_type, home_warehouse_id, status, created_at, empty_return_until
		 FROM equipment ORDER BY unit_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Equipment
	for rows.Next() {
		e := &models.Equipment{}
		var equipType, status string
		if err := rows.Scan(&e.ID, &e.UnitID, &equipType, &e.HomeWarehouseID, &status, &e.CreatedAt, &e.EmptyReturnUntil); err != nil {
			return nil, err
		}
		e.EquipmentType = models.EquipmentType(equipType)
		e.Status = models.EquipmentStatus(status)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *EquipmentRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status models.EquipmentStatus) error {
	_, err := r.db.Exec(ctx,
		`UPDATE equipment SET status=$2, empty_return_until=NULL WHERE id=$1`, id, string(status))
	return err
}

func (r *EquipmentRepo) SetEmptyReturnUntil(ctx context.Context, id uuid.UUID, until time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE equipment SET empty_return_until=$2 WHERE id=$1`, id, until)
	return err
}

func (r *EquipmentRepo) CreateMaintenanceRecord(ctx context.Context, rec *models.MaintenanceRecord) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO maintenance_record (id, equipment_id, description, scheduled_at, estimated_return, completed_at)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		rec.ID, rec.EquipmentID, rec.Description, rec.ScheduledAt, rec.EstimatedReturn, rec.CompletedAt)
	return err
}

func (r *EquipmentRepo) GetActiveMaintenanceByEquipment(ctx context.Context, equipmentID uuid.UUID) (*models.MaintenanceRecord, error) {
	rec := &models.MaintenanceRecord{}
	err := r.db.QueryRow(ctx,
		`SELECT id, equipment_id, description, scheduled_at, estimated_return, completed_at
		 FROM maintenance_record WHERE equipment_id=$1 AND completed_at IS NULL
		 ORDER BY scheduled_at DESC LIMIT 1`, equipmentID).
		Scan(&rec.ID, &rec.EquipmentID, &rec.Description, &rec.ScheduledAt, &rec.EstimatedReturn, &rec.CompletedAt)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

func (r *EquipmentRepo) ResolveMaintenanceRecord(ctx context.Context, id uuid.UUID, completedAt time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE maintenance_record SET completed_at=$2 WHERE id=$1`, id, completedAt)
	return err
}

func (r *EquipmentRepo) CreateBreakdownRecord(ctx context.Context, rec *models.BreakdownRecord) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO breakdown_record (id, equipment_id, breakdown_type, location_desc, driver_id, load_attached, reported_at, resolved_at, tow_cost)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		rec.ID, rec.EquipmentID, string(rec.BreakdownType), rec.LocationDesc, rec.DriverID, rec.LoadAttached, rec.ReportedAt, rec.ResolvedAt, rec.TowCost)
	return err
}

func (r *EquipmentRepo) GetActiveBreakdownByEquipment(ctx context.Context, equipmentID uuid.UUID) (*models.BreakdownRecord, error) {
	rec := &models.BreakdownRecord{}
	var bdType string
	err := r.db.QueryRow(ctx,
		`SELECT id, equipment_id, breakdown_type, location_desc, driver_id, load_attached, reported_at, resolved_at, tow_cost
		 FROM breakdown_record WHERE equipment_id=$1 AND resolved_at IS NULL
		 ORDER BY reported_at DESC LIMIT 1`, equipmentID).
		Scan(&rec.ID, &rec.EquipmentID, &bdType, &rec.LocationDesc, &rec.DriverID, &rec.LoadAttached, &rec.ReportedAt, &rec.ResolvedAt, &rec.TowCost)
	if err != nil {
		return nil, err
	}
	rec.BreakdownType = models.BreakdownType(bdType)
	return rec, nil
}

func (r *EquipmentRepo) ResolveBreakdownRecord(ctx context.Context, id uuid.UUID, resolvedAt time.Time, towCost *float64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE breakdown_record SET resolved_at=$2, tow_cost=$3 WHERE id=$1`,
		id, resolvedAt, towCost)
	return err
}
