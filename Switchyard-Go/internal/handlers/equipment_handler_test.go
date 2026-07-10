package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JacobJGalloway/switchyard-go/internal/events"
	"github.com/JacobJGalloway/switchyard-go/internal/models"
)

// --- minimal stubs ---

type stubEquipRepo struct {
	equipment   *models.Equipment
	maintenance *models.MaintenanceRecord
	breakdown   *models.BreakdownRecord
	createErr   error
	statusErr   error
	maintErr    error
	breakErr    error
	getAllErr    error
	resolveErr  error // used by ResolveMaintenanceRecord and ResolveBreakdownRecord
}

func (r *stubEquipRepo) GetAll(_ context.Context) ([]*models.Equipment, error) {
	return nil, r.getAllErr
}
func (r *stubEquipRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.Equipment, error) {
	if r.equipment == nil {
		return nil, errNotFound
	}
	return r.equipment, nil
}
func (r *stubEquipRepo) Create(_ context.Context, _ *models.Equipment) error  { return r.createErr }
func (r *stubEquipRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ models.EquipmentStatus) error {
	return r.statusErr
}
func (r *stubEquipRepo) CreateMaintenanceRecord(_ context.Context, _ *models.MaintenanceRecord) error {
	return r.maintErr
}
func (r *stubEquipRepo) ResolveMaintenanceRecord(_ context.Context, _ uuid.UUID, _ time.Time) error {
	return r.resolveErr
}
func (r *stubEquipRepo) GetActiveMaintenanceByEquipment(_ context.Context, _ uuid.UUID) (*models.MaintenanceRecord, error) {
	return r.maintenance, nil
}
func (r *stubEquipRepo) CreateBreakdownRecord(_ context.Context, _ *models.BreakdownRecord) error {
	return r.breakErr
}
func (r *stubEquipRepo) ResolveBreakdownRecord(_ context.Context, _ uuid.UUID, _ time.Time, _ *float64) error {
	return r.resolveErr
}
func (r *stubEquipRepo) GetActiveBreakdownByEquipment(_ context.Context, _ uuid.UUID) (*models.BreakdownRecord, error) {
	return r.breakdown, nil
}
func (r *stubEquipRepo) SetEmptyReturnUntil(_ context.Context, _ uuid.UUID, _ time.Time) error {
	return nil
}

type stubEquipNotifier struct{ called bool }

func (n *stubEquipNotifier) OnRoadsideBreakdownWithLoad(_ context.Context, _ events.EquipmentBreakdownPayload) error {
	n.called = true
	return nil
}

// errNotFound is a sentinel used when stubs have no data.
var errNotFound = errNotFoundSentinel{}

type errNotFoundSentinel struct{}

func (e errNotFoundSentinel) Error() string { return "not found" }

// withIDParam injects a chi URL param into the request context.
func withIDParam(r *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func postBody(t testing.TB, body any) *bytes.Reader {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	return bytes.NewReader(raw)
}

// --- GetAll ---

func TestEquipmentGetAll_Returns200(t *testing.T) {
	h := NewEquipmentHandler(&stubEquipRepo{}, &stubEquipNotifier{})
	req := httptest.NewRequest(http.MethodGet, "/api/equipment", nil)
	rec := httptest.NewRecorder()
	h.GetAll(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestEquipmentGetAll_RepoError_Returns500(t *testing.T) {
	h := NewEquipmentHandler(&stubEquipRepo{getAllErr: errNotFound}, &stubEquipNotifier{})
	req := httptest.NewRequest(http.MethodGet, "/api/equipment", nil)
	rec := httptest.NewRecorder()
	h.GetAll(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// --- Create ---

func TestEquipmentCreate_Success(t *testing.T) {
	h := NewEquipmentHandler(&stubEquipRepo{}, &stubEquipNotifier{})
	body := map[string]any{
		"unit_id":          "TRUCK-01",
		"equipment_type":   "truck",
		"home_warehouse_id": "wh-1",
	}
	req := httptest.NewRequest(http.MethodPost, "/api/equipment", postBody(t, body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestEquipmentCreate_MissingUnitID_Returns400(t *testing.T) {
	h := NewEquipmentHandler(&stubEquipRepo{}, &stubEquipNotifier{})
	body := map[string]any{"equipment_type": "truck", "home_warehouse_id": "wh-1"}
	req := httptest.NewRequest(http.MethodPost, "/api/equipment", postBody(t, body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEquipmentCreate_InvalidType_Returns400(t *testing.T) {
	h := NewEquipmentHandler(&stubEquipRepo{}, &stubEquipNotifier{})
	body := map[string]any{"unit_id": "TRUCK-99", "equipment_type": "bicycle", "home_warehouse_id": "wh-1"}
	req := httptest.NewRequest(http.MethodPost, "/api/equipment", postBody(t, body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- ReportMaintenance ---

func TestEquipmentMaintenance_InvalidBody_Returns400(t *testing.T) {
	h := NewEquipmentHandler(&stubEquipRepo{}, &stubEquipNotifier{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader([]byte("not-json"))), uuid.New().String())
	rec := httptest.NewRecorder()
	h.ReportMaintenance(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEquipmentMaintenance_BadUUID_Returns400(t *testing.T) {
	h := NewEquipmentHandler(&stubEquipRepo{}, &stubEquipNotifier{})
	body := map[string]any{"description": "oil change"}
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", postBody(t, body)), "not-a-uuid")
	rec := httptest.NewRecorder()
	h.ReportMaintenance(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEquipmentMaintenance_NotFound_Returns404(t *testing.T) {
	h := NewEquipmentHandler(&stubEquipRepo{}, &stubEquipNotifier{})
	body := map[string]any{"description": "oil change"}
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", postBody(t, body)), uuid.New().String())
	rec := httptest.NewRecorder()
	h.ReportMaintenance(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestEquipmentMaintenance_Success_Returns200(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip}, &stubEquipNotifier{})
	body := map[string]any{"description": "oil change", "scheduled_at": time.Now()}
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", postBody(t, body)), equip.ID.String())
	rec := httptest.NewRecorder()
	h.ReportMaintenance(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestEquipmentMaintenance_AlreadyInMaintenance_Returns409(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusMaintenance}
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip}, &stubEquipNotifier{})
	body := map[string]any{"description": "oil change", "scheduled_at": time.Now()}
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", postBody(t, body)), equip.ID.String())
	rec := httptest.NewRecorder()
	h.ReportMaintenance(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestEquipmentMaintenance_AlreadyInBreakdown_Returns409(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusBreakdown}
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip}, &stubEquipNotifier{})
	body := map[string]any{"description": "flat tire", "scheduled_at": time.Now()}
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", postBody(t, body)), equip.ID.String())
	rec := httptest.NewRecorder()
	h.ReportMaintenance(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestEquipmentMaintenance_MissingDescription_Returns400(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip}, &stubEquipNotifier{})
	body := map[string]any{"scheduled_at": time.Now()}
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", postBody(t, body)), equip.ID.String())
	rec := httptest.NewRecorder()
	h.ReportMaintenance(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- ReportBreakdown ---

func TestEquipmentBreakdown_RoadsideMissingLocationDesc_Returns400(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip}, &stubEquipNotifier{})
	body := map[string]any{"breakdown_type": "roadside", "load_attached": false}
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", postBody(t, body)), equip.ID.String())
	rec := httptest.NewRecorder()
	h.ReportBreakdown(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEquipmentBreakdown_InvalidType_Returns400(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip}, &stubEquipNotifier{})
	body := map[string]any{"breakdown_type": "alien_abduction"}
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", postBody(t, body)), equip.ID.String())
	rec := httptest.NewRecorder()
	h.ReportBreakdown(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEquipmentBreakdown_RoadsideWithLoad_NotifiesDispatcher(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	notifier := &stubEquipNotifier{}
	loc := "I-90 mm 42"
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip}, notifier)
	body := map[string]any{"breakdown_type": "roadside", "load_attached": true, "location_desc": loc}
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", postBody(t, body)), equip.ID.String())
	rec := httptest.NewRecorder()
	h.ReportBreakdown(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, notifier.called, "roadside+load breakdown must alert dispatcher")
}

func TestEquipmentBreakdown_DepotNoLoad_NoNotification(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	notifier := &stubEquipNotifier{}
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip}, notifier)
	body := map[string]any{"breakdown_type": "depot", "load_attached": false}
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", postBody(t, body)), equip.ID.String())
	rec := httptest.NewRecorder()
	h.ReportBreakdown(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.False(t, notifier.called, "depot breakdown must not trigger immediate alert")
}

// --- Resolve ---

func TestEquipmentResolve_BadUUID_Returns400(t *testing.T) {
	h := NewEquipmentHandler(&stubEquipRepo{}, &stubEquipNotifier{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), "not-a-uuid")
	rec := httptest.NewRecorder()
	h.Resolve(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEquipmentResolve_NotFound_Returns404(t *testing.T) {
	h := NewEquipmentHandler(&stubEquipRepo{}, &stubEquipNotifier{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), uuid.New().String())
	rec := httptest.NewRecorder()
	h.Resolve(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestEquipmentResolve_BreakdownRecord_Returns204(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusBreakdown}
	br := &models.BreakdownRecord{ID: uuid.New(), EquipmentID: equip.ID}
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip, breakdown: br}, &stubEquipNotifier{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), equip.ID.String())
	rec := httptest.NewRecorder()
	h.Resolve(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestEquipmentResolve_NotInMaintenanceOrBreakdown_Returns409(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip}, &stubEquipNotifier{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), equip.ID.String())
	rec := httptest.NewRecorder()
	h.Resolve(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestEquipmentResolve_MaintenanceRecord_Returns204(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusMaintenance}
	maint := &models.MaintenanceRecord{ID: uuid.New(), EquipmentID: equip.ID}
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip, maintenance: maint}, &stubEquipNotifier{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), equip.ID.String())
	rec := httptest.NewRecorder()
	h.Resolve(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestEquipmentResolve_MaintenanceNoActiveRecord_Returns404(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusMaintenance}
	// maintenance field is nil — GetActiveMaintenanceByEquipment returns (nil, nil)
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip}, &stubEquipNotifier{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), equip.ID.String())
	rec := httptest.NewRecorder()
	h.Resolve(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestEquipmentResolve_BreakdownNoActiveRecord_Returns404(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusBreakdown}
	// breakdown field is nil — GetActiveBreakdownByEquipment returns (nil, nil)
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip}, &stubEquipNotifier{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), equip.ID.String())
	rec := httptest.NewRecorder()
	h.Resolve(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestEquipmentBreakdown_BadUUID_Returns400(t *testing.T) {
	h := NewEquipmentHandler(&stubEquipRepo{}, &stubEquipNotifier{})
	body := map[string]any{"breakdown_type": "depot", "load_attached": false}
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", postBody(t, body)), "not-a-uuid")
	rec := httptest.NewRecorder()
	h.ReportBreakdown(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEquipmentBreakdown_InvalidDriverID_Returns400(t *testing.T) {
	h := NewEquipmentHandler(&stubEquipRepo{}, &stubEquipNotifier{})
	loc := "dock 3"
	body := map[string]any{
		"breakdown_type": "roadside",
		"location_desc":  loc,
		"driver_id":      "not-a-uuid",
	}
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", postBody(t, body)), uuid.New().String())
	rec := httptest.NewRecorder()
	h.ReportBreakdown(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEquipmentBreakdown_RepoError_Returns500(t *testing.T) {
	h := NewEquipmentHandler(&stubEquipRepo{breakErr: errNotFound}, &stubEquipNotifier{})
	body := map[string]any{"breakdown_type": "depot", "load_attached": false}
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", postBody(t, body)), uuid.New().String())
	rec := httptest.NewRecorder()
	h.ReportBreakdown(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestEquipmentCreate_InvalidBody_Returns400(t *testing.T) {
	h := NewEquipmentHandler(&stubEquipRepo{}, &stubEquipNotifier{})
	req := httptest.NewRequest(http.MethodPost, "/api/equipment", bytes.NewReader([]byte("not-json")))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEquipmentCreate_RepoError_Returns500(t *testing.T) {
	h := NewEquipmentHandler(&stubEquipRepo{createErr: errNotFound}, &stubEquipNotifier{})
	body := map[string]any{"unit_id": "TRUCK-99", "equipment_type": "truck", "home_warehouse_id": "wh-1"}
	req := httptest.NewRequest(http.MethodPost, "/api/equipment", postBody(t, body))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestEquipmentMaintenance_CreateRecordError_Returns500(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip, maintErr: errNotFound}, &stubEquipNotifier{})
	body := map[string]any{"description": "brake check", "scheduled_at": time.Now()}
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", postBody(t, body)), equip.ID.String())
	rec := httptest.NewRecorder()
	h.ReportMaintenance(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestEquipmentMaintenance_UpdateStatusError_Returns500(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusAvailable}
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip, statusErr: errNotFound}, &stubEquipNotifier{})
	body := map[string]any{"description": "brake check", "scheduled_at": time.Now()}
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", postBody(t, body)), equip.ID.String())
	rec := httptest.NewRecorder()
	h.ReportMaintenance(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestEquipmentBreakdown_InvalidBody_Returns400(t *testing.T) {
	h := NewEquipmentHandler(&stubEquipRepo{}, &stubEquipNotifier{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader([]byte("not-json"))), uuid.New().String())
	rec := httptest.NewRecorder()
	h.ReportBreakdown(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestEquipmentBreakdown_UpdateStatusError_Returns500(t *testing.T) {
	h := NewEquipmentHandler(&stubEquipRepo{statusErr: errNotFound}, &stubEquipNotifier{})
	body := map[string]any{"breakdown_type": "depot", "load_attached": false}
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", postBody(t, body)), uuid.New().String())
	rec := httptest.NewRecorder()
	h.ReportBreakdown(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestEquipmentBreakdown_ValidDriverID_Returns200(t *testing.T) {
	// Exercises the uuid.Parse success path for driver_id.
	h := NewEquipmentHandler(&stubEquipRepo{}, &stubEquipNotifier{})
	driverID := uuid.New().String()
	body := map[string]any{"breakdown_type": "depot", "load_attached": false, "driver_id": driverID}
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", postBody(t, body)), uuid.New().String())
	rec := httptest.NewRecorder()
	h.ReportBreakdown(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestEquipmentResolve_MaintenanceResolveError_Returns500(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusMaintenance}
	maint := &models.MaintenanceRecord{ID: uuid.New(), EquipmentID: equip.ID}
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip, maintenance: maint, resolveErr: errNotFound}, &stubEquipNotifier{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), equip.ID.String())
	rec := httptest.NewRecorder()
	h.Resolve(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestEquipmentResolve_BreakdownResolveError_Returns500(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusBreakdown}
	br := &models.BreakdownRecord{ID: uuid.New(), EquipmentID: equip.ID}
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip, breakdown: br, resolveErr: errNotFound}, &stubEquipNotifier{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), equip.ID.String())
	rec := httptest.NewRecorder()
	h.Resolve(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestEquipmentResolve_UpdateStatusError_Returns500(t *testing.T) {
	equip := &models.Equipment{ID: uuid.New(), Status: models.EquipmentStatusMaintenance}
	maint := &models.MaintenanceRecord{ID: uuid.New(), EquipmentID: equip.ID}
	h := NewEquipmentHandler(&stubEquipRepo{equipment: equip, maintenance: maint, statusErr: errNotFound}, &stubEquipNotifier{})
	req := withIDParam(httptest.NewRequest(http.MethodPatch, "/", nil), equip.ID.String())
	rec := httptest.NewRecorder()
	h.Resolve(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
