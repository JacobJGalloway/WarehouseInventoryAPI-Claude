package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JacobJGalloway/switchyard-go/internal/events"
	"github.com/JacobJGalloway/switchyard-go/internal/models"
)

// =============================================================================
// Stub repositories — map-based, return nil/empty when key not found.
// Named wbXxx to avoid conflicts with the mock* types in other test files.
// =============================================================================

type wbDriverRepo struct {
	all  []*models.Driver
	byID map[uuid.UUID]*models.Driver
}

func (r *wbDriverRepo) GetAll(_ context.Context) ([]*models.Driver, error) { return r.all, nil }
func (r *wbDriverRepo) GetByID(_ context.Context, id uuid.UUID) (*models.Driver, error) {
	return r.byID[id], nil
}
func (r *wbDriverRepo) Create(_ context.Context, _ *models.Driver) error { panic("not implemented") }
func (r *wbDriverRepo) Update(_ context.Context, _ *models.Driver) error { panic("not implemented") }

type wbAssignRepo struct {
	active []*models.DriverBOLAssignment
	byID   map[uuid.UUID]*models.DriverBOLAssignment
}

func (r *wbAssignRepo) GetAllActive(_ context.Context) ([]*models.DriverBOLAssignment, error) {
	return r.active, nil
}
func (r *wbAssignRepo) GetByID(_ context.Context, id uuid.UUID) (*models.DriverBOLAssignment, error) {
	return r.byID[id], nil
}
func (r *wbAssignRepo) Create(_ context.Context, _ *models.DriverBOLAssignment) error {
	panic("not implemented")
}
func (r *wbAssignRepo) GetByPlanBOL(_ context.Context, _ uuid.UUID) (*models.DriverBOLAssignment, error) {
	panic("not implemented")
}
func (r *wbAssignRepo) GetActiveByDriver(_ context.Context, _ uuid.UUID) (*models.DriverBOLAssignment, error) {
	panic("not implemented")
}
func (r *wbAssignRepo) MarkDeparted(_ context.Context, _ uuid.UUID, _ time.Time) error {
	panic("not implemented")
}
func (r *wbAssignRepo) MarkFulfilled(_ context.Context, _ uuid.UUID, _ time.Time) error {
	panic("not implemented")
}
func (r *wbAssignRepo) ConfirmDeadhead(_ context.Context, _ uuid.UUID, _ time.Time) error {
	panic("not implemented")
}
func (r *wbAssignRepo) GetCustodyChain(_ context.Context, _ uuid.UUID) ([]*models.DriverBOLAssignment, error) {
	panic("not implemented")
}
func (r *wbAssignRepo) InitiateTransfer(_ context.Context, _ uuid.UUID, _ time.Time, _ uuid.UUID) error {
	panic("not implemented")
}

type wbBOLRepo struct {
	byStatus map[models.PlanBOLStatus][]*models.PlanBOLRecord
	byID     map[uuid.UUID]*models.PlanBOLRecord
	stops    map[uuid.UUID][]*models.PlanBOLStop
	stopsErr error // injected to test GetStops error path in firstUnprocessedStop
}

func (r *wbBOLRepo) GetByStatus(_ context.Context, s models.PlanBOLStatus) ([]*models.PlanBOLRecord, error) {
	return r.byStatus[s], nil
}
func (r *wbBOLRepo) GetByID(_ context.Context, id uuid.UUID) (*models.PlanBOLRecord, error) {
	return r.byID[id], nil
}
func (r *wbBOLRepo) GetStops(_ context.Context, id uuid.UUID) ([]*models.PlanBOLStop, error) {
	return r.stops[id], r.stopsErr
}
func (r *wbBOLRepo) Create(_ context.Context, _ *models.PlanBOLRecord) error {
	panic("not implemented")
}
func (r *wbBOLRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ models.PlanBOLStatus) error {
	panic("not implemented")
}
func (r *wbBOLRepo) SetSubmittedTransactionID(_ context.Context, _ uuid.UUID, _ string) error {
	panic("not implemented")
}
func (r *wbBOLRepo) CreateStop(_ context.Context, _ *models.PlanBOLStop) error {
	panic("not implemented")
}
func (r *wbBOLRepo) GetStopByID(_ context.Context, _ uuid.UUID) (*models.PlanBOLStop, error) {
	panic("not implemented")
}
func (r *wbBOLRepo) MarkStopProcessed(_ context.Context, _ uuid.UUID, _ time.Time, _ *float64) error {
	panic("not implemented")
}
func (r *wbBOLRepo) SetMilesDriven(_ context.Context, _ uuid.UUID, _ float64) error {
	panic("not implemented")
}
func (r *wbBOLRepo) CreateSnapshot(_ context.Context, _ *models.TruckInventorySnapshot) error {
	panic("not implemented")
}
func (r *wbBOLRepo) GetSnapshots(_ context.Context, _ uuid.UUID) ([]*models.TruckInventorySnapshot, error) {
	panic("not implemented")
}
func (r *wbBOLRepo) GetStatusHistory(_ context.Context, _ uuid.UUID) ([]*models.BOLStatusHistory, error) {
	panic("not implemented")
}

type wbEquipRepo struct {
	all         []*models.Equipment
	byID        map[uuid.UUID]*models.Equipment
	maintenance map[uuid.UUID]*models.MaintenanceRecord
	breakdown   map[uuid.UUID]*models.BreakdownRecord
}

func (r *wbEquipRepo) GetAll(_ context.Context) ([]*models.Equipment, error) { return r.all, nil }
func (r *wbEquipRepo) GetByID(_ context.Context, id uuid.UUID) (*models.Equipment, error) {
	return r.byID[id], nil
}
func (r *wbEquipRepo) GetActiveMaintenanceByEquipment(_ context.Context, id uuid.UUID) (*models.MaintenanceRecord, error) {
	return r.maintenance[id], nil
}
func (r *wbEquipRepo) GetActiveBreakdownByEquipment(_ context.Context, id uuid.UUID) (*models.BreakdownRecord, error) {
	return r.breakdown[id], nil
}
func (r *wbEquipRepo) Create(_ context.Context, _ *models.Equipment) error { panic("not implemented") }
func (r *wbEquipRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ models.EquipmentStatus) error {
	panic("not implemented")
}
func (r *wbEquipRepo) CreateMaintenanceRecord(_ context.Context, _ *models.MaintenanceRecord) error {
	panic("not implemented")
}
func (r *wbEquipRepo) ResolveMaintenanceRecord(_ context.Context, _ uuid.UUID, _ time.Time) error {
	panic("not implemented")
}
func (r *wbEquipRepo) CreateBreakdownRecord(_ context.Context, _ *models.BreakdownRecord) error {
	panic("not implemented")
}
func (r *wbEquipRepo) ResolveBreakdownRecord(_ context.Context, _ uuid.UUID, _ time.Time, _ *float64) error {
	panic("not implemented")
}
func (r *wbEquipRepo) SetEmptyReturnUntil(_ context.Context, _ uuid.UUID, _ time.Time) error {
	panic("not implemented")
}

type wbHOSRepo struct {
	windows map[uuid.UUID]*models.HOSWindow
	// key: "stateCode/cycleLabel"
	limits map[string]*models.HOSLimit
}

func (r *wbHOSRepo) GetWindowByDriver(_ context.Context, id uuid.UUID) (*models.HOSWindow, error) {
	return r.windows[id], nil
}
func (r *wbHOSRepo) GetLimitByStateAndCycle(_ context.Context, state, cycle string) (*models.HOSLimit, error) {
	return r.limits[fmt.Sprintf("%s/%s", state, cycle)], nil
}
func (r *wbHOSRepo) CreateLimit(_ context.Context, _ *models.HOSLimit) error { panic("not implemented") }
func (r *wbHOSRepo) CreateWindow(_ context.Context, _ *models.HOSWindow) error {
	panic("not implemented")
}
func (r *wbHOSRepo) UpdateWindow(_ context.Context, _ *models.HOSWindow) error {
	panic("not implemented")
}

type wbPairingRepo struct {
	byActiveBOL map[uuid.UUID]*models.PlanBOLPairing
}

func (r *wbPairingRepo) GetByActiveBOL(_ context.Context, activeBOLID uuid.UUID) (*models.PlanBOLPairing, error) {
	if r == nil {
		return nil, nil
	}
	return r.byActiveBOL[activeBOLID], nil
}
func (r *wbPairingRepo) Create(_ context.Context, _ *models.PlanBOLPairing) error { panic("not implemented") }
func (r *wbPairingRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.PlanBOLPairing, error) {
	panic("not implemented")
}
func (r *wbPairingRepo) GetEligible(_ context.Context, _ string, _ time.Time) ([]*models.PlanBOLPairing, error) {
	panic("not implemented")
}
func (r *wbPairingRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ models.PairingStatus) error {
	panic("not implemented")
}

// newWBService constructs a WhiteboardService wired with the given stubs.
// Default thresholds match the architecture env var defaults. Uses an empty
// pairing repo (no pairings) — use newWBServiceWithPairing for tests that
// need to exercise pairing-aware routing (Empty Return, pairing-at-risk).
func newWBService(dr *wbDriverRepo, ar *wbAssignRepo, br *wbBOLRepo, er *wbEquipRepo, hr *wbHOSRepo) *WhiteboardService {
	return newWBServiceWithPairing(dr, ar, br, er, hr, &wbPairingRepo{byActiveBOL: map[uuid.UUID]*models.PlanBOLPairing{}})
}

func newWBServiceWithPairing(dr *wbDriverRepo, ar *wbAssignRepo, br *wbBOLRepo, er *wbEquipRepo, hr *wbHOSRepo, pr *wbPairingRepo) *WhiteboardService {
	return NewWhiteboardService(dr, ar, br, er, hr,
		2.0,  // warningThresholdHours
		2.0,  // deadheadSearchWindowHours
		4.0,  // loadingAgeThresholdHours
		"60h/7d",
		pr,
		30.0, // cutoffMinutes
		2.0,  // emptyReturnTransitHours
	)
}

// =============================================================================
// hosStatusForWindow tests
// =============================================================================

func TestHOSStatusForWindow(t *testing.T) {
	driverID := uuid.New()
	limit := &models.HOSLimit{DailyDrivingLimitHours: 11, WeeklyLimitHours: 60}
	hos := &wbHOSRepo{limits: map[string]*models.HOSLimit{"IL/60h/7d": limit}}
	svc := newWBService(&wbDriverRepo{}, &wbAssignRepo{}, &wbBOLRepo{}, &wbEquipRepo{}, hos)

	tests := []struct {
		name       string
		window     *models.HOSWindow
		wantStatus HOSStatus
	}{
		{
			name:       "nil window defaults to green",
			window:     nil,
			wantStatus: HOSStatusGreen,
		},
		{
			name:       "ample hours remaining — green",
			window:     &models.HOSWindow{DriverID: driverID, DailyHoursUsed: 5},
			wantStatus: HOSStatusGreen, // remaining = 6, threshold = 2
		},
		{
			name:       "within warning threshold — yellow",
			window:     &models.HOSWindow{DriverID: driverID, DailyHoursUsed: 9.5},
			wantStatus: HOSStatusYellow, // remaining = 1.5, threshold = 2
		},
		{
			name:       "exactly at limit — red",
			window:     &models.HOSWindow{DriverID: driverID, DailyHoursUsed: 11},
			wantStatus: HOSStatusRed, // remaining = 0
		},
		{
			name:       "over limit — red",
			window:     &models.HOSWindow{DriverID: driverID, DailyHoursUsed: 12},
			wantStatus: HOSStatusRed, // remaining < 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.hosStatusForWindow(context.Background(), tt.window, "IL")
			assert.Equal(t, tt.wantStatus, got)
		})
	}
}

func TestHOSStatusForWindow_MissingLimit(t *testing.T) {
	// No limit configured for this state — service must not panic, defaults to green.
	hos := &wbHOSRepo{limits: map[string]*models.HOSLimit{}}
	svc := newWBService(&wbDriverRepo{}, &wbAssignRepo{}, &wbBOLRepo{}, &wbEquipRepo{}, hos)
	window := &models.HOSWindow{DailyHoursUsed: 10}
	assert.Equal(t, HOSStatusGreen, svc.hosStatusForWindow(context.Background(), window, "XX"))
}

// =============================================================================
// computeRestEnd tests (pure function)
// =============================================================================

func TestComputeRestEnd(t *testing.T) {
	stoppedAt := time.Now().UTC()

	t.Run("nil MandatedStopAt returns zero time", func(t *testing.T) {
		window := &models.HOSWindow{}
		end, restType := computeRestEnd(window, nil)
		assert.True(t, end.IsZero())
		assert.Empty(t, restType)
	})

	t.Run("nil limit falls back to 10-hour daily rest", func(t *testing.T) {
		window := &models.HOSWindow{MandatedStopAt: &stoppedAt}
		end, restType := computeRestEnd(window, nil)
		assert.Equal(t, "daily", restType)
		assert.WithinDuration(t, stoppedAt.Add(10*time.Hour), end, time.Second)
	})

	t.Run("weekly hours under limit — daily rest applies", func(t *testing.T) {
		limit := &models.HOSLimit{
			RestPeriodHours:  10,
			WeeklyLimitHours: 60,
			WeeklyResetHours: 34,
		}
		window := &models.HOSWindow{MandatedStopAt: &stoppedAt, WeeklyHoursUsed: 40}
		end, restType := computeRestEnd(window, limit)
		assert.Equal(t, "daily", restType)
		assert.WithinDuration(t, stoppedAt.Add(10*time.Hour), end, time.Second)
	})

	t.Run("weekly hours at limit, weekly reset longer — weekly rest applies", func(t *testing.T) {
		limit := &models.HOSLimit{
			RestPeriodHours:  10,
			WeeklyLimitHours: 60,
			WeeklyResetHours: 34,
		}
		window := &models.HOSWindow{MandatedStopAt: &stoppedAt, WeeklyHoursUsed: 60}
		end, restType := computeRestEnd(window, limit)
		assert.Equal(t, "weekly", restType)
		assert.WithinDuration(t, stoppedAt.Add(34*time.Hour), end, time.Second)
	})

	t.Run("weekly hours at limit but weekly reset shorter than daily — daily wins", func(t *testing.T) {
		limit := &models.HOSLimit{
			RestPeriodHours:  10,
			WeeklyLimitHours: 60,
			WeeklyResetHours: 8, // shorter than daily rest
		}
		window := &models.HOSWindow{MandatedStopAt: &stoppedAt, WeeklyHoursUsed: 60}
		end, restType := computeRestEnd(window, limit)
		assert.Equal(t, "daily", restType)
		assert.WithinDuration(t, stoppedAt.Add(10*time.Hour), end, time.Second)
	})
}

// =============================================================================
// GetBoardState — column routing rules
// =============================================================================

// emptyRepos returns stubs with no data — used as a clean base for each test.
func emptyRepos() (*wbDriverRepo, *wbAssignRepo, *wbBOLRepo, *wbEquipRepo, *wbHOSRepo) {
	return &wbDriverRepo{byID: map[uuid.UUID]*models.Driver{}},
		&wbAssignRepo{byID: map[uuid.UUID]*models.DriverBOLAssignment{}},
		&wbBOLRepo{
			byStatus: map[models.PlanBOLStatus][]*models.PlanBOLRecord{},
			byID:     map[uuid.UUID]*models.PlanBOLRecord{},
			stops:    map[uuid.UUID][]*models.PlanBOLStop{},
		},
		&wbEquipRepo{
			byID:        map[uuid.UUID]*models.Equipment{},
			maintenance: map[uuid.UUID]*models.MaintenanceRecord{},
			breakdown:   map[uuid.UUID]*models.BreakdownRecord{},
		},
		&wbHOSRepo{
			windows: map[uuid.UUID]*models.HOSWindow{},
			limits:  map[string]*models.HOSLimit{},
		}
}

func TestGetBoardState_DraftBOL(t *testing.T) {
	dr, ar, br, er, hr := emptyRepos()
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusDraft, CreatedAt: time.Now()}
	br.byStatus[models.PlanBOLStatusDraft] = []*models.PlanBOLRecord{bol}

	board, err := newWBService(dr, ar, br, er, hr).GetBoardState(context.Background())
	require.NoError(t, err)
	require.Len(t, board.Draft, 1)
	assert.Equal(t, bol.ID, board.Draft[0].PlanBOL.ID)
	assert.Empty(t, board.Pending)
	assert.Empty(t, board.LoadingReady)
}

func TestGetBoardState_PendingBOL(t *testing.T) {
	dr, ar, br, er, hr := emptyRepos()
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusPlanProgress, CreatedAt: time.Now()}
	br.byStatus[models.PlanBOLStatusPlanProgress] = []*models.PlanBOLRecord{bol}

	board, err := newWBService(dr, ar, br, er, hr).GetBoardState(context.Background())
	require.NoError(t, err)
	require.Len(t, board.Pending, 1)
	assert.Equal(t, bol.ID, board.Pending[0].PlanBOL.ID)
}

func TestGetBoardState_LoadingReadyPhases(t *testing.T) {
	dr, ar, br, er, hr := emptyRepos()
	loadingBOL := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusLoading, CreatedAt: time.Now()}
	readyBOL := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusValidated, CreatedAt: time.Now()}
	br.byStatus[models.PlanBOLStatusLoading] = []*models.PlanBOLRecord{loadingBOL}
	br.byStatus[models.PlanBOLStatusValidated] = []*models.PlanBOLRecord{readyBOL}

	board, err := newWBService(dr, ar, br, er, hr).GetBoardState(context.Background())
	require.NoError(t, err)
	require.Len(t, board.LoadingReady, 2)

	var loadingCard, readyCard *LoadingReadyCard
	for _, c := range board.LoadingReady {
		if c.PlanBOL.ID == loadingBOL.ID {
			loadingCard = c
		} else {
			readyCard = c
		}
	}
	require.NotNil(t, loadingCard)
	assert.False(t, loadingCard.IsReady, "loading-status BOL should have IsReady=false")
	require.NotNil(t, readyCard)
	assert.True(t, readyCard.IsReady, "validated-status BOL should have IsReady=true")
}

func TestGetBoardState_LongWaitFlag(t *testing.T) {
	dr, ar, br, er, hr := emptyRepos()
	// BOL created 5 hours ago, threshold is 4 hours → IsLongWait should be set.
	oldBOL := &models.PlanBOLRecord{
		ID:        uuid.New(),
		Status:    models.PlanBOLStatusValidated,
		CreatedAt: time.Now().Add(-5 * time.Hour),
	}
	recentBOL := &models.PlanBOLRecord{
		ID:        uuid.New(),
		Status:    models.PlanBOLStatusValidated,
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}
	br.byStatus[models.PlanBOLStatusValidated] = []*models.PlanBOLRecord{oldBOL, recentBOL}

	board, err := newWBService(dr, ar, br, er, hr).GetBoardState(context.Background())
	require.NoError(t, err)
	require.Len(t, board.LoadingReady, 2)

	for _, c := range board.LoadingReady {
		if c.PlanBOL.ID == oldBOL.ID {
			assert.True(t, c.IsLongWait, "5-hour-old BOL should be flagged as long wait")
		} else {
			assert.False(t, c.IsLongWait, "1-hour-old BOL should not be flagged")
		}
	}
}

func TestGetBoardState_AssignedNotDeparted_StaysInLoadingReady(t *testing.T) {
	// Assigned but DepartedAt=nil — BOL stays in LoadingReady with driver overlay.
	driverID := uuid.New()
	bolID := uuid.New()
	equipID := uuid.New()
	assignID := uuid.New()

	dr, ar, br, er, hr := emptyRepos()
	driver := &models.Driver{ID: driverID, Name: "Alex", LicenseState: "IL", IsActive: true}
	bol := &models.PlanBOLRecord{ID: bolID, Status: models.PlanBOLStatusValidated, CreatedAt: time.Now()}
	assign := &models.DriverBOLAssignment{
		ID: assignID, DriverID: driverID, PlanBOLID: bolID, EquipmentID: equipID,
		AssignedAt: time.Now(),
		// DepartedAt intentionally nil
	}

	dr.byID[driverID] = driver
	dr.all = []*models.Driver{driver}
	ar.active = []*models.DriverBOLAssignment{assign}
	br.byStatus[models.PlanBOLStatusValidated] = []*models.PlanBOLRecord{bol}
	br.byID[bolID] = bol

	board, err := newWBService(dr, ar, br, er, hr).GetBoardState(context.Background())
	require.NoError(t, err)

	require.Len(t, board.LoadingReady, 1)
	card := board.LoadingReady[0]
	require.NotNil(t, card.Assignment, "assignment should be overlaid on the LoadingReady card")
	assert.Equal(t, assignID, card.Assignment.ID)
	assert.NotNil(t, card.Driver)

	assert.Empty(t, board.InDelivery.InTransit, "not-departed assignment must not appear in InDelivery")
}

func TestGetBoardState_DepartedNotFulfilled_InTransit(t *testing.T) {
	driverID := uuid.New()
	bolID := uuid.New()
	equipID := uuid.New()
	now := time.Now()

	dr, ar, br, er, hr := emptyRepos()
	driver := &models.Driver{ID: driverID, Name: "Sam", LicenseState: "IL", IsActive: true}
	bol := &models.PlanBOLRecord{ID: bolID, Status: models.PlanBOLStatusSubmitted, CreatedAt: now}
	equip := &models.Equipment{ID: equipID, UnitID: "TRUCK-01", Status: models.EquipmentStatusAssigned}
	departed := now.Add(-2 * time.Hour)
	assign := &models.DriverBOLAssignment{
		ID: uuid.New(), DriverID: driverID, PlanBOLID: bolID, EquipmentID: equipID,
		AssignedAt: now.Add(-3 * time.Hour),
		DepartedAt: &departed,
		// FulfilledAt nil — still in transit
	}

	dr.byID[driverID] = driver
	ar.active = []*models.DriverBOLAssignment{assign}
	br.byID[bolID] = bol
	er.byID[equipID] = equip
	hr.windows[driverID] = &models.HOSWindow{DriverID: driverID, DailyHoursUsed: 2}

	board, err := newWBService(dr, ar, br, er, hr).GetBoardState(context.Background())
	require.NoError(t, err)

	require.Len(t, board.InDelivery.InTransit, 1)
	card := board.InDelivery.InTransit[0]
	assert.Equal(t, driverID, card.Driver.ID)
	assert.Equal(t, bolID, card.PlanBOL.ID)
	assert.Empty(t, board.Delivered)
	assert.Empty(t, board.InDelivery.MandatedStop)
}

func TestGetBoardState_MandatedStop_SeparateFromInTransit(t *testing.T) {
	driverID := uuid.New()
	bolID := uuid.New()
	equipID := uuid.New()
	now := time.Now()
	departed := now.Add(-4 * time.Hour)
	stoppedAt := now.Add(-1 * time.Hour)

	dr, ar, br, er, hr := emptyRepos()
	dr.byID[driverID] = &models.Driver{ID: driverID, Name: "Casey", LicenseState: "IL", IsActive: true}
	br.byID[bolID] = &models.PlanBOLRecord{ID: bolID, Status: models.PlanBOLStatusSubmitted}
	er.byID[equipID] = &models.Equipment{ID: equipID, UnitID: "TRUCK-02"}
	ar.active = []*models.DriverBOLAssignment{{
		ID: uuid.New(), DriverID: driverID, PlanBOLID: bolID, EquipmentID: equipID,
		AssignedAt: now.Add(-5 * time.Hour),
		DepartedAt: &departed,
	}}
	hr.windows[driverID] = &models.HOSWindow{
		DriverID:       driverID,
		DailyHoursUsed: 9,
		MandatedStopAt: &stoppedAt, // driver is on mandated rest
	}

	board, err := newWBService(dr, ar, br, er, hr).GetBoardState(context.Background())
	require.NoError(t, err)

	assert.Empty(t, board.InDelivery.InTransit, "mandated-stop driver must not appear in InTransit")
	require.Len(t, board.InDelivery.MandatedStop, 1)
	card := board.InDelivery.MandatedStop[0]
	assert.Equal(t, driverID, card.Driver.ID)
	require.NotNil(t, card.MandatedStopAt)
	assert.Equal(t, stoppedAt.Unix(), card.MandatedStopAt.Unix())
}

func TestGetBoardState_Fulfilled_MovesToDelivered(t *testing.T) {
	driverID := uuid.New()
	bolID := uuid.New()
	equipID := uuid.New()
	now := time.Now()
	departed := now.Add(-3 * time.Hour)
	fulfilled := now.Add(-30 * time.Minute)

	dr, ar, br, er, hr := emptyRepos()
	dr.byID[driverID] = &models.Driver{ID: driverID, Name: "Jordan", LicenseState: "IL"}
	br.byID[bolID] = &models.PlanBOLRecord{ID: bolID, Status: models.PlanBOLStatusFulfilled}
	er.byID[equipID] = &models.Equipment{ID: equipID, UnitID: "TRUCK-03"}
	ar.active = []*models.DriverBOLAssignment{{
		ID: uuid.New(), DriverID: driverID, PlanBOLID: bolID, EquipmentID: equipID,
		AssignedAt:  now.Add(-4 * time.Hour),
		DepartedAt:  &departed,
		FulfilledAt: &fulfilled,
	}}

	board, err := newWBService(dr, ar, br, er, hr).GetBoardState(context.Background())
	require.NoError(t, err)

	assert.Empty(t, board.InDelivery.InTransit)
	require.Len(t, board.Delivered, 1, "Delivered is a BOL-only close-out card")
	deliveredCard := board.Delivered[0]
	assert.Equal(t, bolID, deliveredCard.PlanBOL.ID)
	assert.Equal(t, fulfilled.Unix(), deliveredCard.LastStopConfirmedAt.Unix())

	// No pairing secured — driver routes to Empty Return, not Delivered.
	require.Len(t, board.Available.EmptyReturn, 1)
	erCard := board.Available.EmptyReturn[0]
	assert.Equal(t, driverID, erCard.Driver.ID)
	assert.Equal(t, fulfilled.Unix(), erCard.LastStopConfirmedAt.Unix())
	// ETA = fulfilled_at + emptyReturnTransitHours (2h default in newWBService)
	assert.WithinDuration(t, fulfilled.Add(2*time.Hour), erCard.ETA, time.Second)
	// Pairing window = fulfilled_at + deadheadSearchWindow (2h default in newWBService)
	assert.WithinDuration(t, fulfilled.Add(2*time.Hour), erCard.PairingWindowExpiresAt, time.Second)
}

func TestGetBoardState_Fulfilled_PairingSecured_NoEmptyReturn(t *testing.T) {
	driverID := uuid.New()
	bolID := uuid.New()
	equipID := uuid.New()
	now := time.Now()
	departed := now.Add(-3 * time.Hour)
	fulfilled := now.Add(-30 * time.Minute)

	dr, ar, br, er, hr := emptyRepos()
	dr.byID[driverID] = &models.Driver{ID: driverID, Name: "Jordan", LicenseState: "IL"}
	br.byID[bolID] = &models.PlanBOLRecord{ID: bolID, Status: models.PlanBOLStatusFulfilled}
	er.byID[equipID] = &models.Equipment{ID: equipID, UnitID: "TRUCK-03"}
	ar.active = []*models.DriverBOLAssignment{{
		ID: uuid.New(), DriverID: driverID, PlanBOLID: bolID, EquipmentID: equipID,
		AssignedAt:  now.Add(-4 * time.Hour),
		DepartedAt:  &departed,
		FulfilledAt: &fulfilled,
	}}
	pr := &wbPairingRepo{byActiveBOL: map[uuid.UUID]*models.PlanBOLPairing{
		bolID: {ID: uuid.New(), ActiveBOLID: bolID, Status: models.PairingStatusConfirmed},
	}}

	board, err := newWBServiceWithPairing(dr, ar, br, er, hr, pr).GetBoardState(context.Background())
	require.NoError(t, err)

	require.Len(t, board.Delivered, 1, "close-out card still appears regardless of pairing state")
	assert.Empty(t, board.Available.EmptyReturn, "secured pairing routes the driver to the deadhead run, not Empty Return")
}

func TestGetBoardState_PairingAtRisk_NoPairing_SetsFlag(t *testing.T) {
	// cutoffMinutes = 30m (newWBService default). EstimatedRunHours puts estimated
	// fulfillment 10 minutes from now — inside the cutoff window — with no pairing.
	driverID := uuid.New()
	bolID := uuid.New()
	equipID := uuid.New()
	now := time.Now()
	departed := now.Add(-3*time.Hour + 10*time.Minute)
	estimatedRunHours := 3.0

	dr, ar, br, er, hr := emptyRepos()
	dr.byID[driverID] = &models.Driver{ID: driverID, Name: "Riley", LicenseState: "IL"}
	br.byID[bolID] = &models.PlanBOLRecord{ID: bolID, Status: models.PlanBOLStatusSubmitted}
	er.byID[equipID] = &models.Equipment{ID: equipID, UnitID: "TRUCK-20"}
	ar.active = []*models.DriverBOLAssignment{{
		ID: uuid.New(), DriverID: driverID, PlanBOLID: bolID, EquipmentID: equipID,
		AssignedAt:        now.Add(-4 * time.Hour),
		DepartedAt:        &departed,
		EstimatedRunHours: &estimatedRunHours,
	}}

	board, err := newWBService(dr, ar, br, er, hr).GetBoardState(context.Background())
	require.NoError(t, err)

	require.Len(t, board.InDelivery.InTransit, 1)
	assert.True(t, board.InDelivery.InTransit[0].PairingAtRisk)
}

func TestGetBoardState_PairingAtRisk_PairingSecured_NoFlag(t *testing.T) {
	driverID := uuid.New()
	bolID := uuid.New()
	equipID := uuid.New()
	now := time.Now()
	departed := now.Add(-3*time.Hour + 10*time.Minute)
	estimatedRunHours := 3.0

	dr, ar, br, er, hr := emptyRepos()
	dr.byID[driverID] = &models.Driver{ID: driverID, Name: "Riley", LicenseState: "IL"}
	br.byID[bolID] = &models.PlanBOLRecord{ID: bolID, Status: models.PlanBOLStatusSubmitted}
	er.byID[equipID] = &models.Equipment{ID: equipID, UnitID: "TRUCK-20"}
	ar.active = []*models.DriverBOLAssignment{{
		ID: uuid.New(), DriverID: driverID, PlanBOLID: bolID, EquipmentID: equipID,
		AssignedAt:        now.Add(-4 * time.Hour),
		DepartedAt:        &departed,
		EstimatedRunHours: &estimatedRunHours,
	}}
	pr := &wbPairingRepo{byActiveBOL: map[uuid.UUID]*models.PlanBOLPairing{
		bolID: {ID: uuid.New(), ActiveBOLID: bolID, Status: models.PairingStatusProposed},
	}}

	board, err := newWBServiceWithPairing(dr, ar, br, er, hr, pr).GetBoardState(context.Background())
	require.NoError(t, err)

	require.Len(t, board.InDelivery.InTransit, 1)
	assert.False(t, board.InDelivery.InTransit[0].PairingAtRisk)
}

func TestGetBoardState_PairingAtRisk_OutsideCutoffWindow_NoFlag(t *testing.T) {
	// Estimated fulfillment is 2 hours out — well outside the 30m cutoff — so no
	// risk flag even with no pairing.
	driverID := uuid.New()
	bolID := uuid.New()
	equipID := uuid.New()
	now := time.Now()
	departed := now.Add(-1 * time.Hour)
	estimatedRunHours := 3.0

	dr, ar, br, er, hr := emptyRepos()
	dr.byID[driverID] = &models.Driver{ID: driverID, Name: "Riley", LicenseState: "IL"}
	br.byID[bolID] = &models.PlanBOLRecord{ID: bolID, Status: models.PlanBOLStatusSubmitted}
	er.byID[equipID] = &models.Equipment{ID: equipID, UnitID: "TRUCK-20"}
	ar.active = []*models.DriverBOLAssignment{{
		ID: uuid.New(), DriverID: driverID, PlanBOLID: bolID, EquipmentID: equipID,
		AssignedAt:        now.Add(-2 * time.Hour),
		DepartedAt:        &departed,
		EstimatedRunHours: &estimatedRunHours,
	}}

	board, err := newWBService(dr, ar, br, er, hr).GetBoardState(context.Background())
	require.NoError(t, err)

	require.Len(t, board.InDelivery.InTransit, 1)
	assert.False(t, board.InDelivery.InTransit[0].PairingAtRisk)
}

func TestGetBoardState_EquipmentRouting(t *testing.T) {
	// Three equipment items: one in maintenance, one roadside breakdown with load
	// attached (→ InDelivery.Breakdown), one depot breakdown (→ Maintenance).
	maintID := uuid.New()
	roadsideID := uuid.New()
	depotID := uuid.New()

	dr, ar, br, er, hr := emptyRepos()

	er.all = []*models.Equipment{
		{ID: maintID, UnitID: "TRUCK-10", Status: models.EquipmentStatusMaintenance},
		{ID: roadsideID, UnitID: "TRUCK-11", Status: models.EquipmentStatusBreakdown},
		{ID: depotID, UnitID: "TRUCK-12", Status: models.EquipmentStatusBreakdown},
	}
	er.maintenance[maintID] = &models.MaintenanceRecord{EquipmentID: maintID}
	er.breakdown[roadsideID] = &models.BreakdownRecord{
		EquipmentID:   roadsideID,
		BreakdownType: models.BreakdownTypeRoadside,
		LoadAttached:  true,
	}
	er.breakdown[depotID] = &models.BreakdownRecord{
		EquipmentID:   depotID,
		BreakdownType: models.BreakdownTypeDepot,
		LoadAttached:  false,
	}

	board, err := newWBService(dr, ar, br, er, hr).GetBoardState(context.Background())
	require.NoError(t, err)

	// Roadside + load → InDelivery.Breakdown
	require.Len(t, board.InDelivery.Breakdown, 1)
	assert.Equal(t, roadsideID, board.InDelivery.Breakdown[0].Equipment.ID)

	// Maintenance + depot breakdown → Maintenance column
	require.Len(t, board.Maintenance, 2)
	maintenanceIDs := []uuid.UUID{board.Maintenance[0].Equipment.ID, board.Maintenance[1].Equipment.ID}
	assert.Contains(t, maintenanceIDs, maintID)
	assert.Contains(t, maintenanceIDs, depotID)
}

func TestGetBoardState_DriverPool(t *testing.T) {
	// Three drivers: one on active run (skipped), one available, one on mandated rest.
	activeDriverID := uuid.New()
	availDriverID := uuid.New()
	restingDriverID := uuid.New()
	bolID := uuid.New()
	equipID := uuid.New()
	now := time.Now()
	departed := now.Add(-2 * time.Hour)
	stoppedAt := now.Add(-30 * time.Minute)

	dr, ar, br, er, hr := emptyRepos()
	dr.all = []*models.Driver{
		{ID: activeDriverID, Name: "Active", LicenseState: "IL", IsActive: true},
		{ID: availDriverID, Name: "Available", LicenseState: "IL", IsActive: true},
		{ID: restingDriverID, Name: "Resting", LicenseState: "IL", IsActive: true},
	}
	for _, d := range dr.all {
		dr.byID[d.ID] = d
	}
	ar.active = []*models.DriverBOLAssignment{{
		ID: uuid.New(), DriverID: activeDriverID, PlanBOLID: bolID, EquipmentID: equipID,
		AssignedAt: now.Add(-3 * time.Hour),
		DepartedAt: &departed,
	}}
	br.byID[bolID] = &models.PlanBOLRecord{ID: bolID, Status: models.PlanBOLStatusSubmitted}
	er.byID[equipID] = &models.Equipment{ID: equipID}
	hr.windows[activeDriverID] = &models.HOSWindow{DriverID: activeDriverID}
	hr.windows[availDriverID] = &models.HOSWindow{DriverID: availDriverID, DailyHoursUsed: 3}
	hr.windows[restingDriverID] = &models.HOSWindow{
		DriverID:       restingDriverID,
		DailyHoursUsed: 10,
		MandatedStopAt: &stoppedAt,
	}

	board, err := newWBService(dr, ar, br, er, hr).GetBoardState(context.Background())
	require.NoError(t, err)

	availIDs := make([]uuid.UUID, len(board.Available.AvailableNow))
	for i, c := range board.Available.AvailableNow {
		availIDs[i] = c.Driver.ID
	}
	assert.Contains(t, availIDs, availDriverID)
	assert.NotContains(t, availIDs, activeDriverID, "on-run driver must not appear in Available")
	assert.NotContains(t, availIDs, restingDriverID, "resting driver must not appear in AvailableNow")

	require.Len(t, board.Available.Resting, 1)
	assert.Equal(t, restingDriverID, board.Available.Resting[0].Driver.ID)
}

// =============================================================================
// GetBoardState — firstUnprocessedStop (sort + loop body)
// =============================================================================

func TestGetBoardState_InTransit_PicksFirstUnprocessedStop(t *testing.T) {
	driverID := uuid.New()
	bolID := uuid.New()
	equipID := uuid.New()
	now := time.Now()
	departed := now.Add(-2 * time.Hour)

	dr, ar, br, er, hr := emptyRepos()
	driver := &models.Driver{ID: driverID, Name: "River", LicenseState: "IL", IsActive: true}
	bol := &models.PlanBOLRecord{ID: bolID, Status: models.PlanBOLStatusSubmitted, CreatedAt: now}
	equip := &models.Equipment{ID: equipID}
	assign := &models.DriverBOLAssignment{
		ID: uuid.New(), DriverID: driverID, PlanBOLID: bolID, EquipmentID: equipID,
		AssignedAt: now.Add(-3 * time.Hour), DepartedAt: &departed,
	}

	stop1 := &models.PlanBOLStop{ID: uuid.New(), PlanBOLID: bolID, Sequence: 1, IsProcessed: true}
	stop2 := &models.PlanBOLStop{ID: uuid.New(), PlanBOLID: bolID, Sequence: 2, IsProcessed: false}

	dr.byID[driverID] = driver
	ar.active = []*models.DriverBOLAssignment{assign}
	br.byID[bolID] = bol
	br.stops[bolID] = []*models.PlanBOLStop{stop2, stop1} // out-of-order to exercise sort
	er.byID[equipID] = equip
	hr.windows[driverID] = &models.HOSWindow{DriverID: driverID, DailyHoursUsed: 1}

	board, err := newWBService(dr, ar, br, er, hr).GetBoardState(context.Background())
	require.NoError(t, err)
	require.Len(t, board.InDelivery.InTransit, 1)
	require.NotNil(t, board.InDelivery.InTransit[0].CurrentStop)
	assert.Equal(t, stop2.ID, board.InDelivery.InTransit[0].CurrentStop.ID, "first unprocessed stop should be seq=2")
}

func TestGetBoardState_InTransit_StopsRepoError_NilCurrentStop(t *testing.T) {
	// GetStops returns an error → firstUnprocessedStop returns nil defensively.
	driverID := uuid.New()
	bolID := uuid.New()
	equipID := uuid.New()
	now := time.Now()
	departed := now.Add(-2 * time.Hour)

	dr, ar, br, er, hr := emptyRepos()
	br.stopsErr = fmt.Errorf("db timeout")
	dr.byID[driverID] = &models.Driver{ID: driverID, Name: "Chris", LicenseState: "IL", IsActive: true}
	br.byID[bolID] = &models.PlanBOLRecord{ID: bolID, Status: models.PlanBOLStatusSubmitted}
	er.byID[equipID] = &models.Equipment{ID: equipID}
	ar.active = []*models.DriverBOLAssignment{{
		ID: uuid.New(), DriverID: driverID, PlanBOLID: bolID, EquipmentID: equipID,
		AssignedAt: now.Add(-3 * time.Hour), DepartedAt: &departed,
	}}
	hr.windows[driverID] = &models.HOSWindow{DriverID: driverID, DailyHoursUsed: 2}

	board, err := newWBService(dr, ar, br, er, hr).GetBoardState(context.Background())
	require.NoError(t, err)
	require.Len(t, board.InDelivery.InTransit, 1)
	assert.Nil(t, board.InDelivery.InTransit[0].CurrentStop, "GetStops error → nil current stop")
}

func TestGetBoardState_InTransit_AllStopsProcessed_NilCurrentStop(t *testing.T) {
	// firstUnprocessedStop returns nil when every stop on the BOL is processed.
	driverID := uuid.New()
	bolID := uuid.New()
	equipID := uuid.New()
	now := time.Now()
	departed := now.Add(-2 * time.Hour)

	dr, ar, br, er, hr := emptyRepos()
	dr.byID[driverID] = &models.Driver{ID: driverID, Name: "Pat", LicenseState: "IL", IsActive: true}
	br.byID[bolID] = &models.PlanBOLRecord{ID: bolID, Status: models.PlanBOLStatusSubmitted}
	er.byID[equipID] = &models.Equipment{ID: equipID}
	ar.active = []*models.DriverBOLAssignment{{
		ID: uuid.New(), DriverID: driverID, PlanBOLID: bolID, EquipmentID: equipID,
		AssignedAt: now.Add(-3 * time.Hour), DepartedAt: &departed,
	}}
	// All stops marked processed — firstUnprocessedStop returns nil.
	br.stops[bolID] = []*models.PlanBOLStop{
		{ID: uuid.New(), PlanBOLID: bolID, Sequence: 1, IsProcessed: true},
		{ID: uuid.New(), PlanBOLID: bolID, Sequence: 2, IsProcessed: true},
	}
	hr.windows[driverID] = &models.HOSWindow{DriverID: driverID, DailyHoursUsed: 2}

	board, err := newWBService(dr, ar, br, er, hr).GetBoardState(context.Background())
	require.NoError(t, err)
	require.Len(t, board.InDelivery.InTransit, 1)
	assert.Nil(t, board.InDelivery.InTransit[0].CurrentStop, "all stops processed — no current stop")
}

// =============================================================================
// GetAlerts — alert generation branches
// =============================================================================

func TestGetAlerts_EmptyBoard_ReturnsNilError(t *testing.T) {
	dr, ar, br, er, hr := emptyRepos()
	svc := newWBService(dr, ar, br, er, hr)
	alerts, err := svc.GetAlerts(context.Background())
	require.NoError(t, err)
	assert.Empty(t, alerts)
}

func TestGetAlerts_HOSWarning_GeneratesAlert(t *testing.T) {
	driverID := uuid.New()
	bolID := uuid.New()
	equipID := uuid.New()
	now := time.Now()
	departed := now.Add(-2 * time.Hour)

	dr, ar, br, er, hr := emptyRepos()
	limit := &models.HOSLimit{DailyDrivingLimitHours: 11, WeeklyLimitHours: 60}
	hr.limits["IL/60h/7d"] = limit

	driver := &models.Driver{ID: driverID, Name: "Dana", LicenseState: "IL", IsActive: true}
	bol := &models.PlanBOLRecord{ID: bolID, Status: models.PlanBOLStatusSubmitted, CreatedAt: now}
	equip := &models.Equipment{ID: equipID}
	assign := &models.DriverBOLAssignment{
		ID: uuid.New(), DriverID: driverID, PlanBOLID: bolID, EquipmentID: equipID,
		AssignedAt: now.Add(-3 * time.Hour), DepartedAt: &departed,
	}

	// 9.5 hours used → remaining = 1.5, threshold = 2 → HOSStatusYellow
	dr.byID[driverID] = driver
	ar.active = []*models.DriverBOLAssignment{assign}
	br.byID[bolID] = bol
	er.byID[equipID] = equip
	hr.windows[driverID] = &models.HOSWindow{DriverID: driverID, DailyHoursUsed: 9.5}

	alerts, err := newWBService(dr, ar, br, er, hr).GetAlerts(context.Background())
	require.NoError(t, err)
	require.Len(t, alerts, 1)
	assert.Equal(t, AlertTypeHOSWarning, alerts[0].AlertType)
}

func TestGetAlerts_RestingDriver_GeneratesHOSWeeklyAlert(t *testing.T) {
	driverID := uuid.New()
	now := time.Now()
	stoppedAt := now.Add(-1 * time.Hour)

	dr, ar, br, er, hr := emptyRepos()
	driver := &models.Driver{ID: driverID, Name: "Rested", LicenseState: "IL", IsActive: true}
	dr.all = []*models.Driver{driver}
	dr.byID[driverID] = driver
	// driver is NOT in any active assignment — no entries in ar.active
	hr.windows[driverID] = &models.HOSWindow{
		DriverID:       driverID,
		DailyHoursUsed: 10,
		MandatedStopAt: &stoppedAt,
	}
	// need limit so computeRestEnd can compute reset duration
	hr.limits["IL/60h/7d"] = &models.HOSLimit{
		DailyDrivingLimitHours: 11,
		WeeklyLimitHours:       60,
		RestPeriodHours:        10,
		WeeklyResetHours:       34,
	}

	alerts, err := newWBService(dr, ar, br, er, hr).GetAlerts(context.Background())
	require.NoError(t, err)
	require.Len(t, alerts, 1)
	assert.Equal(t, AlertTypeHOSWeeklyLimit, alerts[0].AlertType)
	require.NotNil(t, alerts[0].DriverID)
	assert.Equal(t, driverID, *alerts[0].DriverID)
}

func TestGetAlerts_RoadsideBreakdownWithDriver_GeneratesCriticalAlert(t *testing.T) {
	equipID := uuid.New()
	driverID := uuid.New()

	dr, ar, br, er, hr := emptyRepos()
	driver := &models.Driver{ID: driverID, Name: "Marcus", LicenseState: "OH", IsActive: true}
	dr.byID[driverID] = driver

	er.all = []*models.Equipment{
		{ID: equipID, UnitID: "TK-101", Status: models.EquipmentStatusBreakdown},
	}
	er.byID[equipID] = er.all[0]
	er.breakdown[equipID] = &models.BreakdownRecord{
		EquipmentID:   equipID,
		BreakdownType: models.BreakdownTypeRoadside,
		LoadAttached:  true,
		DriverID:      &driverID,
	}

	alerts, err := newWBService(dr, ar, br, er, hr).GetAlerts(context.Background())
	require.NoError(t, err)
	require.Len(t, alerts, 1)
	assert.Equal(t, AlertTypeRoadsideBreakdown, alerts[0].AlertType)
	assert.Equal(t, AlertSeverityCritical, alerts[0].Severity)
	require.NotNil(t, alerts[0].DriverID)
	assert.Equal(t, driverID, *alerts[0].DriverID)
	assert.Contains(t, alerts[0].Message, "TK-101")
	assert.Contains(t, alerts[0].Message, "Marcus")
}

func TestGetAlerts_RoadsideBreakdownWithoutDriver_GeneratesCriticalAlert(t *testing.T) {
	equipID := uuid.New()

	dr, ar, br, er, hr := emptyRepos()
	er.all = []*models.Equipment{
		{ID: equipID, UnitID: "TK-102", Status: models.EquipmentStatusBreakdown},
	}
	er.byID[equipID] = er.all[0]
	er.breakdown[equipID] = &models.BreakdownRecord{
		EquipmentID:   equipID,
		BreakdownType: models.BreakdownTypeRoadside,
		LoadAttached:  true,
		DriverID:      nil,
	}

	alerts, err := newWBService(dr, ar, br, er, hr).GetAlerts(context.Background())
	require.NoError(t, err)
	require.Len(t, alerts, 1)
	assert.Equal(t, AlertTypeRoadsideBreakdown, alerts[0].AlertType)
	assert.Nil(t, alerts[0].DriverID)
	assert.Contains(t, alerts[0].Message, "TK-102")
}

func TestGetAlerts_ExpiringDeadhead_GeneratesAlert(t *testing.T) {
	// deadheadSearchWindow = 2h (newWBService default)
	// expiringThreshold = 1h
	// FulfilledAt 1.5h ago → expiresAt = now+0.5h → remaining ≈ 30m ≤ 1h → alert
	driverID := uuid.New()
	bolID := uuid.New()
	equipID := uuid.New()
	now := time.Now()
	departed := now.Add(-3 * time.Hour)
	fulfilled := now.Add(-90 * time.Minute)

	dr, ar, br, er, hr := emptyRepos()
	dr.byID[driverID] = &models.Driver{ID: driverID, Name: "Jordan", LicenseState: "IL"}
	br.byID[bolID] = &models.PlanBOLRecord{ID: bolID, Status: models.PlanBOLStatusFulfilled}
	er.byID[equipID] = &models.Equipment{ID: equipID, UnitID: "TK-50"}
	assignID := uuid.New()
	ar.active = []*models.DriverBOLAssignment{{
		ID:          assignID,
		DriverID:    driverID,
		PlanBOLID:   bolID,
		EquipmentID: equipID,
		AssignedAt:  now.Add(-4 * time.Hour),
		DepartedAt:  &departed,
		FulfilledAt: &fulfilled,
	}}

	// No pairing secured — driver routes to Empty Return, and this alert now
	// reads from Available.EmptyReturn's PairingWindowExpiresAt rather than
	// the old Delivered-card countdown (which no longer carries a driver at all).
	alerts, err := newWBService(dr, ar, br, er, hr).GetAlerts(context.Background())
	require.NoError(t, err)
	require.Len(t, alerts, 1)
	assert.Equal(t, AlertTypeExpiringDeadhead, alerts[0].AlertType)
	require.NotNil(t, alerts[0].DriverID)
	assert.Equal(t, driverID, *alerts[0].DriverID)
}

func TestGetAlerts_NonExpiringDeadhead_NoAlert(t *testing.T) {
	// FulfilledAt 30m ago → expiresAt = now+1.5h → remaining > 1h → no alert
	driverID := uuid.New()
	bolID := uuid.New()
	equipID := uuid.New()
	now := time.Now()
	departed := now.Add(-2 * time.Hour)
	fulfilled := now.Add(-30 * time.Minute)

	dr, ar, br, er, hr := emptyRepos()
	dr.byID[driverID] = &models.Driver{ID: driverID, Name: "Alex", LicenseState: "IL"}
	br.byID[bolID] = &models.PlanBOLRecord{ID: bolID, Status: models.PlanBOLStatusFulfilled}
	er.byID[equipID] = &models.Equipment{ID: equipID}
	ar.active = []*models.DriverBOLAssignment{{
		ID:          uuid.New(),
		DriverID:    driverID,
		PlanBOLID:   bolID,
		EquipmentID: equipID,
		AssignedAt:  now.Add(-3 * time.Hour),
		DepartedAt:  &departed,
		FulfilledAt: &fulfilled,
	}}

	alerts, err := newWBService(dr, ar, br, er, hr).GetAlerts(context.Background())
	require.NoError(t, err)
	assert.Empty(t, alerts)
}

// =============================================================================
// On* callbacks — v1.1 no-ops, all must return nil
// =============================================================================

func TestWhiteboardService_OnCallbacks_ReturnNil(t *testing.T) {
	svc := newWBService(&wbDriverRepo{}, &wbAssignRepo{}, &wbBOLRepo{}, &wbEquipRepo{}, &wbHOSRepo{})
	ctx := context.Background()
	assert.NoError(t, svc.OnAssignmentDeparted(ctx, events.AssignmentPayload{}))
	assert.NoError(t, svc.OnAssignmentFulfilled(ctx, events.AssignmentPayload{}))
	assert.NoError(t, svc.OnDeadheadConfirmed(ctx, events.AssignmentPayload{}))
	assert.NoError(t, svc.OnMandatedStop(ctx, events.MandatedStopPayload{}))
	assert.NoError(t, svc.OnEquipmentBreakdown(ctx, events.EquipmentBreakdownPayload{}))
	assert.NoError(t, svc.OnEquipmentResolved(ctx, events.EquipmentResolvedPayload{}))
}
