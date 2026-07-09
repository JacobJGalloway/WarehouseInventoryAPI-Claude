package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JacobJGalloway/switchyard-go/internal/events"
	"github.com/JacobJGalloway/switchyard-go/internal/models"
)

// --- minimal stubs ---

type stubAssignRepo struct {
	assignment         *models.DriverBOLAssignment
	planBOLAssignment  *models.DriverBOLAssignment
	custodyChain       []*models.DriverBOLAssignment
	createErr          error
	markDepartedErr    error
	markFulfilledErr   error
	confirmDeadheadErr error
	initiateTransferErr error
}

func (r *stubAssignRepo) GetAllActive(_ context.Context) ([]*models.DriverBOLAssignment, error) {
	return nil, nil
}
func (r *stubAssignRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.DriverBOLAssignment, error) {
	if r.assignment == nil {
		return nil, errNotFound
	}
	return r.assignment, nil
}
func (r *stubAssignRepo) Create(_ context.Context, _ *models.DriverBOLAssignment) error {
	return r.createErr
}
func (r *stubAssignRepo) GetByPlanBOL(_ context.Context, _ uuid.UUID) (*models.DriverBOLAssignment, error) {
	if r.planBOLAssignment == nil {
		return nil, errNotFound
	}
	return r.planBOLAssignment, nil
}
func (r *stubAssignRepo) GetActiveByDriver(_ context.Context, _ uuid.UUID) (*models.DriverBOLAssignment, error) {
	return nil, nil
}
func (r *stubAssignRepo) MarkDeparted(_ context.Context, _ uuid.UUID, _ time.Time) error {
	return r.markDepartedErr
}
func (r *stubAssignRepo) MarkFulfilled(_ context.Context, _ uuid.UUID, _ time.Time) error {
	return r.markFulfilledErr
}
func (r *stubAssignRepo) ConfirmDeadhead(_ context.Context, _ uuid.UUID, _ time.Time) error {
	return r.confirmDeadheadErr
}
func (r *stubAssignRepo) GetCustodyChain(_ context.Context, _ uuid.UUID) ([]*models.DriverBOLAssignment, error) {
	return r.custodyChain, nil
}
func (r *stubAssignRepo) InitiateTransfer(_ context.Context, _ uuid.UUID, _ time.Time, _ uuid.UUID) error {
	return r.initiateTransferErr
}

type stubDriverRepo struct {
	driver    *models.Driver
	all       []*models.Driver
	err       error
	getAllErr  error
}

func (r *stubDriverRepo) GetAll(_ context.Context) ([]*models.Driver, error) {
	return r.all, r.getAllErr
}
func (r *stubDriverRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.Driver, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.driver, nil
}
func (r *stubDriverRepo) Create(_ context.Context, _ *models.Driver) error { return nil }
func (r *stubDriverRepo) Update(_ context.Context, _ *models.Driver) error { return nil }

type stubBOLRepo struct {
	bol                   *models.PlanBOLRecord
	stop                  *models.PlanBOLStop
	stopErr               error
	stops                 []*models.PlanBOLStop
	stopsErr              error
	updateStatusErr       error
	createStopErr         error
	markStopErr           error
	statusHistoryErr      error
	setTransactionIDErr   error
}

func (r *stubBOLRepo) Create(_ context.Context, _ *models.PlanBOLRecord) error { return nil }
func (r *stubBOLRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.PlanBOLRecord, error) {
	if r.bol == nil {
		return nil, errNotFound
	}
	return r.bol, nil
}
func (r *stubBOLRepo) GetByStatus(_ context.Context, _ models.PlanBOLStatus) ([]*models.PlanBOLRecord, error) {
	return nil, nil
}
func (r *stubBOLRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ models.PlanBOLStatus) error {
	return r.updateStatusErr
}
func (r *stubBOLRepo) SetSubmittedTransactionID(_ context.Context, _ uuid.UUID, _ string) error {
	return r.setTransactionIDErr
}
func (r *stubBOLRepo) CreateStop(_ context.Context, _ *models.PlanBOLStop) error {
	return r.createStopErr
}
func (r *stubBOLRepo) GetStops(_ context.Context, _ uuid.UUID) ([]*models.PlanBOLStop, error) {
	return r.stops, r.stopsErr
}
func (r *stubBOLRepo) GetStopByID(_ context.Context, _ uuid.UUID) (*models.PlanBOLStop, error) {
	if r.stopErr != nil {
		return nil, r.stopErr
	}
	return r.stop, nil
}
func (r *stubBOLRepo) MarkStopProcessed(_ context.Context, _ uuid.UUID, _ time.Time, _ *float64) error {
	return r.markStopErr
}
func (r *stubBOLRepo) SetMilesDriven(_ context.Context, _ uuid.UUID, _ float64) error { return nil }
func (r *stubBOLRepo) CreateSnapshot(_ context.Context, _ *models.TruckInventorySnapshot) error {
	return nil
}
func (r *stubBOLRepo) GetSnapshots(_ context.Context, _ uuid.UUID) ([]*models.TruckInventorySnapshot, error) {
	return nil, nil
}
func (r *stubBOLRepo) GetStatusHistory(_ context.Context, _ uuid.UUID) ([]*models.BOLStatusHistory, error) {
	return nil, r.statusHistoryErr
}

type stubAssignEquipRepo struct {
	equipment            *models.Equipment
	updateStatusErr      error
	setEmptyReturnErr    error
	updateStatusCalls    []models.EquipmentStatus
	setEmptyReturnCalled bool
}

func (r *stubAssignEquipRepo) GetAll(_ context.Context) ([]*models.Equipment, error) { return nil, nil }
func (r *stubAssignEquipRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.Equipment, error) {
	if r.equipment == nil {
		return nil, errNotFound
	}
	return r.equipment, nil
}
func (r *stubAssignEquipRepo) Create(_ context.Context, _ *models.Equipment) error { return nil }
func (r *stubAssignEquipRepo) UpdateStatus(_ context.Context, _ uuid.UUID, status models.EquipmentStatus) error {
	r.updateStatusCalls = append(r.updateStatusCalls, status)
	return r.updateStatusErr
}
func (r *stubAssignEquipRepo) CreateMaintenanceRecord(_ context.Context, _ *models.MaintenanceRecord) error {
	return nil
}
func (r *stubAssignEquipRepo) ResolveMaintenanceRecord(_ context.Context, _ uuid.UUID, _ time.Time) error {
	return nil
}
func (r *stubAssignEquipRepo) GetActiveMaintenanceByEquipment(_ context.Context, _ uuid.UUID) (*models.MaintenanceRecord, error) {
	return nil, nil
}
func (r *stubAssignEquipRepo) CreateBreakdownRecord(_ context.Context, _ *models.BreakdownRecord) error {
	return nil
}
func (r *stubAssignEquipRepo) ResolveBreakdownRecord(_ context.Context, _ uuid.UUID, _ time.Time, _ *float64) error {
	return nil
}
func (r *stubAssignEquipRepo) GetActiveBreakdownByEquipment(_ context.Context, _ uuid.UUID) (*models.BreakdownRecord, error) {
	return nil, nil
}
func (r *stubAssignEquipRepo) SetEmptyReturnUntil(_ context.Context, _ uuid.UUID, _ time.Time) error {
	r.setEmptyReturnCalled = true
	return r.setEmptyReturnErr
}

type stubAssignPairingRepo struct {
	pairing *models.PlanBOLPairing
}

func (r *stubAssignPairingRepo) Create(_ context.Context, _ *models.PlanBOLPairing) error {
	return nil
}
func (r *stubAssignPairingRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.PlanBOLPairing, error) {
	return r.pairing, nil
}
func (r *stubAssignPairingRepo) GetByActiveBOL(_ context.Context, _ uuid.UUID) (*models.PlanBOLPairing, error) {
	return r.pairing, nil
}
func (r *stubAssignPairingRepo) GetEligible(_ context.Context, _ string, _ time.Time) ([]*models.PlanBOLPairing, error) {
	return nil, nil
}
func (r *stubAssignPairingRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ models.PairingStatus) error {
	return nil
}

type stubHOSSvc struct{ err error }

func (s *stubHOSSvc) CanAssign(_ context.Context, _ uuid.UUID, _ float64, _, _ string) error {
	return s.err
}

type stubWBSvc struct{}

func (s *stubWBSvc) OnAssignmentDeparted(_ context.Context, _ events.AssignmentPayload) error {
	return nil
}
func (s *stubWBSvc) OnAssignmentFulfilled(_ context.Context, _ events.AssignmentPayload) error {
	return nil
}
func (s *stubWBSvc) OnDeadheadConfirmed(_ context.Context, _ events.AssignmentPayload) error {
	return nil
}

type stubAssignNotifySvc struct{}

func (s *stubAssignNotifySvc) OnBOLWorkflowCompleted(_ context.Context, _ events.BOLCompletedPayload) error {
	return nil
}

func newAssignHandler(
	assign *stubAssignRepo,
	driver *stubDriverRepo,
	bol *stubBOLRepo,
	equip *stubAssignEquipRepo,
	hos *stubHOSSvc,
) *AssignmentHandler {
	return newAssignHandlerWithPairing(assign, driver, bol, equip, hos, &stubAssignPairingRepo{})
}

func newAssignHandlerWithPairing(
	assign *stubAssignRepo,
	driver *stubDriverRepo,
	bol *stubBOLRepo,
	equip *stubAssignEquipRepo,
	hos *stubHOSSvc,
	pairing *stubAssignPairingRepo,
) *AssignmentHandler {
	return NewAssignmentHandler(assign, driver, bol, equip, hos, &stubWBSvc{}, &stubAssignNotifySvc{}, pairing, 2.0)
}

// --- Get ---

func TestAssignmentGet_BadUUID_Returns400(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodGet, "/", nil), "not-a-uuid")
	rec := httptest.NewRecorder()
	h.Get(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAssignmentGet_NotFound_Returns404(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodGet, "/", nil), uuid.New().String())
	rec := httptest.NewRecorder()
	h.Get(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAssignmentGet_Success_Returns200(t *testing.T) {
	assign := &models.DriverBOLAssignment{ID: uuid.New(), DriverID: uuid.New(), PlanBOLID: uuid.New()}
	h := newAssignHandler(&stubAssignRepo{assignment: assign}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodGet, "/", nil), assign.ID.String())
	rec := httptest.NewRecorder()
	h.Get(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- Create ---

func TestAssignmentCreate_BOLNotValidated_Returns409(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusLoading}
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{equipment: equip}, &stubHOSSvc{})
	body := map[string]any{
		"driver_id":    uuid.New().String(),
		"plan_bol_id":  bol.ID.String(),
		"equipment_id": equip.ID.String(),
		"state_code":   "IL",
		"cycle_label":  "60h/7d",
	}
	req := httptest.NewRequest(http.MethodPost, "/api/assignment", postBody(t, body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestAssignmentCreate_EquipmentNotAvailable_Returns409(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusValidated}
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAssigned}
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{equipment: equip}, &stubHOSSvc{})
	body := map[string]any{
		"driver_id":    uuid.New().String(),
		"plan_bol_id":  bol.ID.String(),
		"equipment_id": equip.ID.String(),
		"state_code":   "IL",
		"cycle_label":  "60h/7d",
	}
	req := httptest.NewRequest(http.MethodPost, "/api/assignment", postBody(t, body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestAssignmentCreate_EmptyReturnWindowElapsed_ReconciledToAvailable(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusValidated}
	past := time.Now().Add(-1 * time.Minute)
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAssigned, EmptyReturnUntil: &past}
	equipRepo := &stubAssignEquipRepo{equipment: equip}
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, equipRepo, &stubHOSSvc{})
	body := map[string]any{
		"driver_id":    uuid.New().String(),
		"plan_bol_id":  bol.ID.String(),
		"equipment_id": equip.ID.String(),
		"state_code":   "IL",
		"cycle_label":  "60h/7d",
	}
	req := httptest.NewRequest(http.MethodPost, "/api/assignment", postBody(t, body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code, "elapsed empty-return window should reconcile equipment to Available and allow the new assignment")
}

func TestAssignmentCreate_EmptyReturnWindowNotYetElapsed_Returns409(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusValidated}
	future := time.Now().Add(1 * time.Hour)
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAssigned, EmptyReturnUntil: &future}
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{equipment: equip}, &stubHOSSvc{})
	body := map[string]any{
		"driver_id":    uuid.New().String(),
		"plan_bol_id":  bol.ID.String(),
		"equipment_id": equip.ID.String(),
		"state_code":   "IL",
		"cycle_label":  "60h/7d",
	}
	req := httptest.NewRequest(http.MethodPost, "/api/assignment", postBody(t, body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code, "equipment still within its empty-return window should not be assignable")
}

func TestAssignmentCreate_HOSViolation_Returns422(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusValidated}
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	hos := &stubHOSSvc{err: errNotFound} // any non-nil error simulates HOS rejection
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{equipment: equip}, hos)
	body := map[string]any{
		"driver_id":           uuid.New().String(),
		"plan_bol_id":         bol.ID.String(),
		"equipment_id":        equip.ID.String(),
		"state_code":          "IL",
		"cycle_label":         "60h/7d",
		"estimated_run_hours": 12.0,
	}
	req := httptest.NewRequest(http.MethodPost, "/api/assignment", postBody(t, body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAssignmentCreate_NilBody_Returns400(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := httptest.NewRequest(http.MethodPost, "/api/assignment", nil)
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAssignmentCreate_InvalidDriverID_Returns400(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	body := map[string]any{
		"driver_id":    "not-a-uuid",
		"plan_bol_id":  uuid.New().String(),
		"equipment_id": uuid.New().String(),
		"state_code":   "IL",
		"cycle_label":  "60h/7d",
	}
	req := httptest.NewRequest(http.MethodPost, "/api/assignment", postBody(t, body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAssignmentCreate_BOLNotFound_Returns404(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	body := map[string]any{
		"driver_id":    uuid.New().String(),
		"plan_bol_id":  uuid.New().String(),
		"equipment_id": uuid.New().String(),
		"state_code":   "IL",
		"cycle_label":  "60h/7d",
	}
	req := httptest.NewRequest(http.MethodPost, "/api/assignment", postBody(t, body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAssignmentCreate_EquipmentNotFound_Returns404(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusValidated}
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	body := map[string]any{
		"driver_id":    uuid.New().String(),
		"plan_bol_id":  bol.ID.String(),
		"equipment_id": uuid.New().String(),
		"state_code":   "IL",
		"cycle_label":  "60h/7d",
	}
	req := httptest.NewRequest(http.MethodPost, "/api/assignment", postBody(t, body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAssignmentCreate_Success_Returns201(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusValidated}
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{equipment: equip}, &stubHOSSvc{})
	body := map[string]any{
		"driver_id":           uuid.New().String(),
		"plan_bol_id":         bol.ID.String(),
		"equipment_id":        equip.ID.String(),
		"state_code":          "IL",
		"cycle_label":         "60h/7d",
		"estimated_run_hours": 4.0,
	}
	req := httptest.NewRequest(http.MethodPost, "/api/assignment", postBody(t, body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestAssignmentCreate_MissingStateCode_Returns400(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusValidated}
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{equipment: equip}, &stubHOSSvc{})
	body := map[string]any{
		"driver_id":    uuid.New().String(),
		"plan_bol_id":  bol.ID.String(),
		"equipment_id": equip.ID.String(),
		// state_code missing
	}
	req := httptest.NewRequest(http.MethodPost, "/api/assignment", postBody(t, body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- Depart ---

func TestAssignmentDepart_BadUUID_Returns400(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), "not-a-uuid")
	rec := httptest.NewRecorder()
	h.Depart(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAssignmentDepart_NotFound_Returns404(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), uuid.New().String())
	rec := httptest.NewRecorder()
	h.Depart(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAssignmentDepart_AlreadyDeparted_Returns409(t *testing.T) {
	departed := time.Now()
	assign := &models.DriverBOLAssignment{ID: uuid.New(), DepartedAt: &departed}
	h := newAssignHandler(&stubAssignRepo{assignment: assign}, &stubDriverRepo{}, &stubBOLRepo{bol: &models.PlanBOLRecord{}}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), assign.ID.String())
	rec := httptest.NewRecorder()
	h.Depart(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestAssignmentDepart_Success_Returns204(t *testing.T) {
	assign := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: uuid.New()}
	h := newAssignHandler(&stubAssignRepo{assignment: assign}, &stubDriverRepo{}, &stubBOLRepo{bol: &models.PlanBOLRecord{}}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), assign.ID.String())
	rec := httptest.NewRecorder()
	h.Depart(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// --- Fulfill ---

func TestAssignmentFulfill_BadUUID_Returns400(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), "not-a-uuid")
	rec := httptest.NewRecorder()
	h.Fulfill(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAssignmentFulfill_NotFound_Returns404(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), uuid.New().String())
	rec := httptest.NewRecorder()
	h.Fulfill(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAssignmentFulfill_AlreadyFulfilled_Returns409(t *testing.T) {
	fulfilledAt := time.Now()
	assign := &models.DriverBOLAssignment{ID: uuid.New(), FulfilledAt: &fulfilledAt}
	h := newAssignHandler(&stubAssignRepo{assignment: assign}, &stubDriverRepo{}, &stubBOLRepo{bol: &models.PlanBOLRecord{}}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), assign.ID.String())
	rec := httptest.NewRecorder()
	h.Fulfill(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestAssignmentFulfill_Success_Returns204(t *testing.T) {
	assign := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: uuid.New()}
	h := newAssignHandler(&stubAssignRepo{assignment: assign}, &stubDriverRepo{}, &stubBOLRepo{bol: &models.PlanBOLRecord{}}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), assign.ID.String())
	rec := httptest.NewRecorder()
	h.Fulfill(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestAssignmentFulfill_NoPairing_SetsEmptyReturnTimer(t *testing.T) {
	assign := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: uuid.New()}
	equip := &stubAssignEquipRepo{}
	h := newAssignHandlerWithPairing(&stubAssignRepo{assignment: assign}, &stubDriverRepo{}, &stubBOLRepo{bol: &models.PlanBOLRecord{}}, equip, &stubHOSSvc{}, &stubAssignPairingRepo{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), assign.ID.String())
	rec := httptest.NewRecorder()
	h.Fulfill(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.True(t, equip.setEmptyReturnCalled, "no pairing secured — equipment should get an empty-return timer, not an immediate release")
	assert.Empty(t, equip.updateStatusCalls)
}

func TestAssignmentFulfill_PairingSecured_ReleasesEquipmentImmediately(t *testing.T) {
	assign := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: uuid.New()}
	equip := &stubAssignEquipRepo{}
	pairing := &stubAssignPairingRepo{pairing: &models.PlanBOLPairing{ID: uuid.New(), Status: models.PairingStatusConfirmed}}
	h := newAssignHandlerWithPairing(&stubAssignRepo{assignment: assign}, &stubDriverRepo{}, &stubBOLRepo{bol: &models.PlanBOLRecord{}}, equip, &stubHOSSvc{}, pairing)
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), assign.ID.String())
	rec := httptest.NewRecorder()
	h.Fulfill(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.False(t, equip.setEmptyReturnCalled, "pairing secured — equipment continues immediately, no empty-return timer")
	require.Len(t, equip.updateStatusCalls, 1)
	assert.Equal(t, models.EquipmentStatusAvailable, equip.updateStatusCalls[0])
}

// --- ConfirmDeadhead ---

func TestAssignmentConfirmDeadhead_BadUUID_Returns400(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), "not-a-uuid")
	rec := httptest.NewRecorder()
	h.ConfirmDeadhead(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAssignmentConfirmDeadhead_NotFound_Returns404(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), uuid.New().String())
	rec := httptest.NewRecorder()
	h.ConfirmDeadhead(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAssignmentConfirmDeadhead_NotYetFulfilledOrTransferred_Returns409(t *testing.T) {
	assign := &models.DriverBOLAssignment{ID: uuid.New(), FulfilledAt: nil, TransferredAt: nil}
	h := newAssignHandler(&stubAssignRepo{assignment: assign}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), assign.ID.String())
	rec := httptest.NewRecorder()
	h.ConfirmDeadhead(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestAssignmentConfirmDeadhead_TransferDeadhead_Returns204(t *testing.T) {
	transferredAt := time.Now().Add(-30 * time.Minute)
	assign := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: uuid.New(), TransferredAt: &transferredAt}
	h := newAssignHandler(&stubAssignRepo{assignment: assign}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), assign.ID.String())
	rec := httptest.NewRecorder()
	h.ConfirmDeadhead(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestAssignmentConfirmDeadhead_AlreadyConfirmed_Returns409(t *testing.T) {
	fulfilledAt := time.Now().Add(-1 * time.Hour)
	confirmedAt := time.Now()
	assign := &models.DriverBOLAssignment{
		ID:                  uuid.New(),
		FulfilledAt:         &fulfilledAt,
		DeadheadConfirmedAt: &confirmedAt,
	}
	h := newAssignHandler(&stubAssignRepo{assignment: assign}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), assign.ID.String())
	rec := httptest.NewRecorder()
	h.ConfirmDeadhead(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestAssignmentConfirmDeadhead_Success_Returns204(t *testing.T) {
	fulfilledAt := time.Now().Add(-1 * time.Hour)
	assign := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: uuid.New(), FulfilledAt: &fulfilledAt}
	h := newAssignHandler(&stubAssignRepo{assignment: assign}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), assign.ID.String())
	rec := httptest.NewRecorder()
	h.ConfirmDeadhead(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestAssignmentConfirmDeadhead_RepoError_Returns500(t *testing.T) {
	fulfilledAt := time.Now().Add(-1 * time.Hour)
	assign := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: uuid.New(), FulfilledAt: &fulfilledAt}
	h := newAssignHandler(&stubAssignRepo{assignment: assign, confirmDeadheadErr: errNotFound}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), assign.ID.String())
	rec := httptest.NewRecorder()
	h.ConfirmDeadhead(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestAssignmentCreate_RepoError_Returns500(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusValidated}
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	h := newAssignHandler(&stubAssignRepo{createErr: errNotFound}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{equipment: equip}, &stubHOSSvc{})
	body := map[string]any{
		"driver_id": uuid.New().String(), "plan_bol_id": bol.ID.String(),
		"equipment_id": equip.ID.String(), "state_code": "IL", "cycle_label": "60h/7d",
	}
	req := httptest.NewRequest(http.MethodPost, "/api/assignment", postBody(t, body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestAssignmentCreate_EquipStatusError_Returns500(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusValidated}
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{equipment: equip, updateStatusErr: errNotFound}, &stubHOSSvc{})
	body := map[string]any{
		"driver_id": uuid.New().String(), "plan_bol_id": bol.ID.String(),
		"equipment_id": equip.ID.String(), "state_code": "IL", "cycle_label": "60h/7d",
	}
	req := httptest.NewRequest(http.MethodPost, "/api/assignment", postBody(t, body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestAssignmentDepart_MarkDepartedError_Returns500(t *testing.T) {
	assign := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: uuid.New()}
	h := newAssignHandler(&stubAssignRepo{assignment: assign, markDepartedErr: errNotFound}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), assign.ID.String())
	rec := httptest.NewRecorder()
	h.Depart(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestAssignmentDepart_UpdateBOLStatusError_Returns500(t *testing.T) {
	assign := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: uuid.New()}
	h := newAssignHandler(&stubAssignRepo{assignment: assign}, &stubDriverRepo{}, &stubBOLRepo{updateStatusErr: errNotFound}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), assign.ID.String())
	rec := httptest.NewRecorder()
	h.Depart(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestAssignmentFulfill_MarkFulfilledError_Returns500(t *testing.T) {
	assign := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: uuid.New()}
	h := newAssignHandler(&stubAssignRepo{assignment: assign, markFulfilledErr: errNotFound}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), assign.ID.String())
	rec := httptest.NewRecorder()
	h.Fulfill(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestAssignmentFulfill_UpdateBOLStatusError_Returns500(t *testing.T) {
	assign := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: uuid.New()}
	h := newAssignHandler(&stubAssignRepo{assignment: assign}, &stubDriverRepo{}, &stubBOLRepo{updateStatusErr: errNotFound}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), assign.ID.String())
	rec := httptest.NewRecorder()
	h.Fulfill(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// --- Transfer ---

func validTransferBody(bolID, driverID, equipID uuid.UUID) map[string]any {
	return map[string]any{
		"incoming_driver_id":    driverID.String(),
		"incoming_equipment_id": equipID.String(),
		"transfer_location_id":  "STOP-IL-47",
		"transfer_reason":       "hos_limit",
		"estimated_run_hours":   3.5,
		"state_code":            "IL",
		"cycle_label":           "60h/7d",
	}
}

func TestTransfer_BadUUID_Returns400(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPost, "/", nil), "not-a-uuid")
	rec := httptest.NewRecorder()
	h.Transfer(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTransfer_NilBody_Returns400(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodPost, "/", nil), uuid.New().String())
	rec := httptest.NewRecorder()
	h.Transfer(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTransfer_InvalidIncomingDriverID_Returns400(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	body := map[string]any{
		"incoming_driver_id":    "not-a-uuid",
		"incoming_equipment_id": uuid.New().String(),
		"transfer_location_id":  "STOP-IL-47",
		"transfer_reason":       "hos_limit",
		"state_code":            "IL",
		"cycle_label":           "60h/7d",
	}
	req := withIDParam(httptest.NewRequest(http.MethodPost, "/", postBody(t, body)), uuid.New().String())
	rec := httptest.NewRecorder()
	h.Transfer(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTransfer_InvalidIncomingEquipmentID_Returns400(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	body := map[string]any{
		"incoming_driver_id":    uuid.New().String(),
		"incoming_equipment_id": "not-a-uuid",
		"transfer_location_id":  "STOP-IL-47",
		"transfer_reason":       "hos_limit",
		"state_code":            "IL",
		"cycle_label":           "60h/7d",
	}
	req := withIDParam(httptest.NewRequest(http.MethodPost, "/", postBody(t, body)), uuid.New().String())
	rec := httptest.NewRecorder()
	h.Transfer(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTransfer_MissingLocationID_Returns400(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	body := map[string]any{
		"incoming_driver_id":    uuid.New().String(),
		"incoming_equipment_id": uuid.New().String(),
		"transfer_reason":       "hos_limit",
		"state_code":            "IL",
		"cycle_label":           "60h/7d",
	}
	req := withIDParam(httptest.NewRequest(http.MethodPost, "/", postBody(t, body)), uuid.New().String())
	rec := httptest.NewRecorder()
	h.Transfer(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTransfer_InvalidTransferReason_Returns400(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	body := map[string]any{
		"incoming_driver_id":    uuid.New().String(),
		"incoming_equipment_id": uuid.New().String(),
		"transfer_location_id":  "STOP-IL-47",
		"transfer_reason":       "bad_reason",
		"state_code":            "IL",
		"cycle_label":           "60h/7d",
	}
	req := withIDParam(httptest.NewRequest(http.MethodPost, "/", postBody(t, body)), uuid.New().String())
	rec := httptest.NewRecorder()
	h.Transfer(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTransfer_MissingStateCode_Returns400(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	body := map[string]any{
		"incoming_driver_id":    uuid.New().String(),
		"incoming_equipment_id": uuid.New().String(),
		"transfer_location_id":  "STOP-IL-47",
		"transfer_reason":       "hos_limit",
		"cycle_label":           "60h/7d",
	}
	req := withIDParam(httptest.NewRequest(http.MethodPost, "/", postBody(t, body)), uuid.New().String())
	rec := httptest.NewRecorder()
	h.Transfer(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTransfer_BOLNotFound_Returns404(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	body := validTransferBody(uuid.New(), uuid.New(), uuid.New())
	req := withIDParam(httptest.NewRequest(http.MethodPost, "/", postBody(t, body)), uuid.New().String())
	rec := httptest.NewRecorder()
	h.Transfer(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestTransfer_BOLNotSubmitted_Returns409(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusValidated}
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	body := validTransferBody(bol.ID, uuid.New(), uuid.New())
	req := withIDParam(httptest.NewRequest(http.MethodPost, "/", postBody(t, body)), bol.ID.String())
	rec := httptest.NewRecorder()
	h.Transfer(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestTransfer_NoActiveAssignment_Returns404(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusSubmitted}
	// planBOLAssignment nil → GetByPlanBOL returns errNotFound
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	body := validTransferBody(bol.ID, uuid.New(), uuid.New())
	req := withIDParam(httptest.NewRequest(http.MethodPost, "/", postBody(t, body)), bol.ID.String())
	rec := httptest.NewRecorder()
	h.Transfer(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestTransfer_NotYetDeparted_Returns409(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusSubmitted}
	current := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: bol.ID, DepartedAt: nil}
	h := newAssignHandler(&stubAssignRepo{planBOLAssignment: current}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	body := validTransferBody(bol.ID, uuid.New(), uuid.New())
	req := withIDParam(httptest.NewRequest(http.MethodPost, "/", postBody(t, body)), bol.ID.String())
	rec := httptest.NewRecorder()
	h.Transfer(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestTransfer_AlreadyFulfilled_Returns409(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusSubmitted}
	departed := time.Now().Add(-2 * time.Hour)
	fulfilled := time.Now().Add(-30 * time.Minute)
	current := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: bol.ID, DepartedAt: &departed, FulfilledAt: &fulfilled}
	h := newAssignHandler(&stubAssignRepo{planBOLAssignment: current}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	body := validTransferBody(bol.ID, uuid.New(), uuid.New())
	req := withIDParam(httptest.NewRequest(http.MethodPost, "/", postBody(t, body)), bol.ID.String())
	rec := httptest.NewRecorder()
	h.Transfer(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestTransfer_IncomingEquipNotFound_Returns404(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusSubmitted}
	departed := time.Now().Add(-2 * time.Hour)
	current := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: bol.ID, DepartedAt: &departed}
	// equipment nil → GetByID returns errNotFound
	h := newAssignHandler(&stubAssignRepo{planBOLAssignment: current}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	body := validTransferBody(bol.ID, uuid.New(), uuid.New())
	req := withIDParam(httptest.NewRequest(http.MethodPost, "/", postBody(t, body)), bol.ID.String())
	rec := httptest.NewRecorder()
	h.Transfer(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestTransfer_IncomingEquipNotAvailable_Returns409(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusSubmitted}
	departed := time.Now().Add(-2 * time.Hour)
	current := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: bol.ID, DepartedAt: &departed}
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAssigned}
	h := newAssignHandler(&stubAssignRepo{planBOLAssignment: current}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{equipment: equip}, &stubHOSSvc{})
	body := validTransferBody(bol.ID, uuid.New(), equip.ID)
	req := withIDParam(httptest.NewRequest(http.MethodPost, "/", postBody(t, body)), bol.ID.String())
	rec := httptest.NewRecorder()
	h.Transfer(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestTransfer_HOSViolation_Returns422(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusSubmitted}
	departed := time.Now().Add(-2 * time.Hour)
	current := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: bol.ID, DepartedAt: &departed}
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	h := newAssignHandler(&stubAssignRepo{planBOLAssignment: current}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{equipment: equip}, &stubHOSSvc{err: errNotFound})
	body := validTransferBody(bol.ID, uuid.New(), equip.ID)
	req := withIDParam(httptest.NewRequest(http.MethodPost, "/", postBody(t, body)), bol.ID.String())
	rec := httptest.NewRecorder()
	h.Transfer(rec, req)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestTransfer_Success_Returns201(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusSubmitted}
	departed := time.Now().Add(-2 * time.Hour)
	current := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: bol.ID, DepartedAt: &departed}
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	h := newAssignHandler(&stubAssignRepo{planBOLAssignment: current}, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{equipment: equip}, &stubHOSSvc{})
	body := validTransferBody(bol.ID, uuid.New(), equip.ID)
	req := withIDParam(httptest.NewRequest(http.MethodPost, "/", postBody(t, body)), bol.ID.String())
	rec := httptest.NewRecorder()
	h.Transfer(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestTransfer_InitiateTransferError_Returns500(t *testing.T) {
	bol := &models.PlanBOLRecord{ID: uuid.New(), Status: models.PlanBOLStatusSubmitted}
	departed := time.Now().Add(-2 * time.Hour)
	current := &models.DriverBOLAssignment{ID: uuid.New(), PlanBOLID: bol.ID, DepartedAt: &departed}
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	assignRepo := &stubAssignRepo{planBOLAssignment: current, initiateTransferErr: errNotFound}
	h := newAssignHandler(assignRepo, &stubDriverRepo{}, &stubBOLRepo{bol: bol}, &stubAssignEquipRepo{equipment: equip}, &stubHOSSvc{})
	body := validTransferBody(bol.ID, uuid.New(), equip.ID)
	req := withIDParam(httptest.NewRequest(http.MethodPost, "/", postBody(t, body)), bol.ID.String())
	rec := httptest.NewRecorder()
	h.Transfer(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// --- GetCustodyChain ---

func TestGetCustodyChain_BadUUID_Returns400(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodGet, "/", nil), "not-a-uuid")
	rec := httptest.NewRecorder()
	h.GetCustodyChain(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetCustodyChain_EmptyChain_Returns200(t *testing.T) {
	h := newAssignHandler(&stubAssignRepo{custodyChain: nil}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodGet, "/", nil), uuid.New().String())
	rec := httptest.NewRecorder()
	h.GetCustodyChain(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetCustodyChain_SingleSegment_Returns200(t *testing.T) {
	chain := []*models.DriverBOLAssignment{
		{ID: uuid.New(), DriverID: uuid.New(), PlanBOLID: uuid.New(), EquipmentID: uuid.New()},
	}
	h := newAssignHandler(&stubAssignRepo{custodyChain: chain}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodGet, "/", nil), chain[0].PlanBOLID.String())
	rec := httptest.NewRecorder()
	h.GetCustodyChain(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetCustodyChain_TwoSegments_Returns200(t *testing.T) {
	bolID := uuid.New()
	chain := []*models.DriverBOLAssignment{
		{ID: uuid.New(), DriverID: uuid.New(), PlanBOLID: bolID, EquipmentID: uuid.New()},
		{ID: uuid.New(), DriverID: uuid.New(), PlanBOLID: bolID, EquipmentID: uuid.New()},
	}
	h := newAssignHandler(&stubAssignRepo{custodyChain: chain}, &stubDriverRepo{}, &stubBOLRepo{}, &stubAssignEquipRepo{}, &stubHOSSvc{})
	req := withIDParam(httptest.NewRequest(http.MethodGet, "/", nil), bolID.String())
	rec := httptest.NewRecorder()
	h.GetCustodyChain(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}
