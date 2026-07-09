package main

import (
	"context"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"

	"github.com/JacobJGalloway/switchyard-go/internal/events"
	"github.com/JacobJGalloway/switchyard-go/internal/handlers"
	"github.com/JacobJGalloway/switchyard-go/internal/integrations"
	authmw "github.com/JacobJGalloway/switchyard-go/internal/middleware"
	pgdb "github.com/JacobJGalloway/switchyard-go/internal/repository/postgres"
	"github.com/JacobJGalloway/switchyard-go/internal/services"
)

func main() {
	viper.AutomaticEnv()
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("SMTP_PORT", "587")
	viper.SetDefault("HOS_WARNING_THRESHOLD_HOURS", 2.0)
	viper.SetDefault("DEADHEAD_WINDOW_HOURS", 4.0)
	viper.SetDefault("DEADHEAD_CUTOFF_MINUTES", 30.0)
	viper.SetDefault("DEADHEAD_SEARCH_WINDOW_HOURS", 2.0)
	viper.SetDefault("LOADING_AGE_THRESHOLD_HOURS", 4.0)
	viper.SetDefault("DEFAULT_CYCLE_LABEL", "60h/7d")
	viper.SetDefault("EMPTY_RETURN_TRANSIT_HOURS", 2.0)
	viper.SetDefault("REACT_BASE_URL", "https://localhost:5173")

	dbURL := viper.GetString("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	readDBURL := viper.GetString("READ_DATABASE_URL")
	if readDBURL == "" {
		log.Fatal("READ_DATABASE_URL is required")
	}

	// --- Migrations ---
	// golang-migrate pgx/v5 driver expects pgx5:// scheme.
	migrateURL := strings.Replace(dbURL, "postgres://", "pgx5://", 1)
	m, err := migrate.New("file://internal/migrations", migrateURL)
	if err != nil {
		log.Fatalf("migrate init: %v", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migrate up: %v", err)
	}
	srcErr, dbErr := m.Close()
	if srcErr != nil {
		log.Printf("migrate close source: %v", srcErr)
	}
	if dbErr != nil {
		log.Printf("migrate close db: %v", dbErr)
	}

	// --- Database pool ---
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("pgxpool: %v", err)
	}
	defer pool.Close()

	// Read replica — logically replicated via native Postgres publication/subscription
	// from a genuinely separate instance (see docker-compose.yml switchyard-pg-read).
	// DR-scoped: isolates read queries from a bad write or corruption on the primary.
	readPool, err := pgxpool.New(context.Background(), readDBURL)
	if err != nil {
		log.Fatalf("read pgxpool: %v", err)
	}
	defer readPool.Close()

	// --- Repositories ---
	driverRepo := pgdb.NewDriverRepo(pool)
	hosRepo := pgdb.NewHOSRepo(pool)
	equipRepo := pgdb.NewEquipmentRepo(pool)
	assignRepo := pgdb.NewAssignmentRepo(pool)
	pairingRepo := pgdb.NewPairingRepo(pool)
	bolRepo := pgdb.NewPlanBOLRepo(pool)
	invoiceRepo := pgdb.NewInvoiceRepo(pool)
	warehouseRepo := pgdb.NewWarehouseRepo(pool)

	// --- Notification service (no upstream deps) ---
	notifySvc := services.NewNotificationService(services.NotificationConfig{
		SMTPHost:      viper.GetString("SMTP_HOST"),
		SMTPPort:      viper.GetString("SMTP_PORT"),
		SMTPUser:      viper.GetString("SMTP_USER"),
		SMTPPass:      viper.GetString("SMTP_PASS"),
		DispatchEmail: viper.GetString("DISPATCH_EMAIL"),
	})

	// --- HOS service ---
	hosSvc := services.NewHOSService(hosRepo, notifySvc, viper.GetFloat64("HOS_WARNING_THRESHOLD_HOURS"))

	// --- Integration clients ---
	// Lazy token provider: the event handler holds and refreshes the M2M token.
	// The closure captures the handler pointer so clients remain usable before
	// the handler is constructed — no request can arrive until ListenAndServe.
	var eventHandler *events.Handler
	tokenProvider := integrations.TokenProvider(func() string {
		if eventHandler == nil {
			return ""
		}
		return eventHandler.TokenProvider()()
	})
	invClient := integrations.NewInventoryClient(viper.GetString("INVENTORY_BASE_URL"), tokenProvider)
	logClient := integrations.NewLogisticsClient(viper.GetString("LOGISTICS_BASE_URL"), tokenProvider)

	// --- Application services ---
	routePlannerSvc := services.NewRoutePlannerService(bolRepo, invClient, hosSvc)
	wbSvc := services.NewWhiteboardService(
		driverRepo,
		assignRepo,
		bolRepo,
		equipRepo,
		hosRepo,
		viper.GetFloat64("HOS_WARNING_THRESHOLD_HOURS"),
		viper.GetFloat64("DEADHEAD_SEARCH_WINDOW_HOURS"),
		viper.GetFloat64("LOADING_AGE_THRESHOLD_HOURS"),
		viper.GetString("DEFAULT_CYCLE_LABEL"),
		pairingRepo,
		viper.GetFloat64("DEADHEAD_CUTOFF_MINUTES"),
		viper.GetFloat64("EMPTY_RETURN_TRANSIT_HOURS"),
	)

	// --- Event handler (sole M2M token owner) ---
	eventHandler = events.NewHandler(
		events.Config{
			Auth0Domain:   viper.GetString("AUTH0_DOMAIN"),
			Auth0ClientID: viper.GetString("AUTH0_CLIENT_ID"),
			Auth0Secret:   viper.GetString("AUTH0_CLIENT_SECRET"),
			Auth0Audience: viper.GetString("AUTH0_AUDIENCE"),
		},
		hosSvc,
		wbSvc,
		notifySvc,
		invClient,
		logClient,
	)

	// --- Templates ---
	tmpl, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		log.Fatalf("parsing templates: %v", err)
	}

	// --- HTTP handlers ---
	planBOLHandler := handlers.NewPlanBOLHandler(routePlannerSvc, bolRepo, logClient)
	driverHandler := handlers.NewDriverHandler(driverRepo, bolRepo, hosRepo, assignRepo, hosSvc, logClient, tmpl)
	assignmentHandler := handlers.NewAssignmentHandler(assignRepo, driverRepo, bolRepo, equipRepo, hosSvc, wbSvc, notifySvc, pairingRepo, viper.GetFloat64("EMPTY_RETURN_TRANSIT_HOURS"))
	equipmentHandler := handlers.NewEquipmentHandler(equipRepo, notifySvc)
	deadheadHandler := handlers.NewDeadheadHandler(pairingRepo, viper.GetFloat64("DEADHEAD_WINDOW_HOURS"), viper.GetFloat64("DEADHEAD_CUTOFF_MINUTES"))
	invoiceHandler := handlers.NewInvoiceHandler(invoiceRepo)
	whiteboardHandler := handlers.NewWhiteboardHandler(wbSvc, tmpl, viper.GetString("REACT_BASE_URL"))
	analyticsRepo := pgdb.NewAnalyticsRepo(readPool)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsRepo)
	regionalInvHandler := handlers.NewRegionalInventoryHandler(invClient, warehouseRepo)

	// --- Router ---
	// --- JWT middleware ---
	checkJWT, err := authmw.NewJWTMiddleware(
		viper.GetString("AUTH0_DOMAIN"),
		viper.GetString("AUTH0_AUDIENCE"),
	)
	if err != nil {
		log.Fatalf("jwt middleware: %v", err)
	}

	// --- Router ---
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Server-rendered pages — browser session auth is a separate concern
	r.Get("/", whiteboardHandler.GetBoardPage)
	r.Get("/driver/{id}", driverHandler.GetRunsheetPage)

	// All /api/* routes require a valid JWT
	r.Group(func(r chi.Router) {
		r.Use(checkJWT)

		// Event ingress (M2M token refreshed transparently by handler)
		r.Post("/api/events", eventHandler.Handle)

		// Dispatch board
		r.Get("/api/dispatch/board", whiteboardHandler.GetBoard)
		r.Get("/api/dispatch/alerts", whiteboardHandler.GetAlerts)

		// Plan BOL — read and most state transitions open to any authenticated user
		r.Post("/api/plan-bol", planBOLHandler.Create)
		r.Get("/api/plan-bol/{id}", planBOLHandler.Get)
		r.Post("/api/plan-bol/{id}/validate", planBOLHandler.Validate)
		r.Patch("/api/plan-bol/{id}/mark-loaded", planBOLHandler.MarkLoaded)
		r.Get("/api/plan-bol/{id}/truck-state", planBOLHandler.GetTruckState)
		r.Get("/api/plan-bol/{id}/history", planBOLHandler.GetStatusHistory)

		// Dispatcher / route planner only — requires bol:plan permission
		r.Group(func(r chi.Router) {
			r.Use(authmw.RequirePermission("plan:bol"))
			r.Post("/api/plan-bol/{id}/begin-planning", planBOLHandler.BeginPlanning)
			r.Post("/api/plan-bol/{id}/commit", planBOLHandler.Commit)
			r.Post("/api/plan-bol/{id}/return-depot", planBOLHandler.AddReturnDepot)
		})

		// Drivers
		r.Get("/api/driver", driverHandler.GetAll)
		r.Get("/api/driver/{id}/runsheet", driverHandler.GetRunsheet)
		r.Get("/api/driver/{id}/active-bol", driverHandler.GetActiveBOL)
		r.Get("/api/driver/{id}/hos", driverHandler.GetHOS)
		r.Post("/api/driver/{id}/stop/{stopId}/log", driverHandler.LogStop)

		// Assignments
		r.Post("/api/assignment", assignmentHandler.Create)
		r.Get("/api/assignment/{id}", assignmentHandler.Get)
		r.Patch("/api/assignment/{id}/depart", assignmentHandler.Depart)
		r.Patch("/api/assignment/{id}/fulfill", assignmentHandler.Fulfill)
		r.Patch("/api/assignment/{id}/deadhead", assignmentHandler.ConfirmDeadhead)
		r.Post("/api/plan-bol/{id}/transfer", assignmentHandler.Transfer)
		r.Get("/api/plan-bol/{id}/assignments", assignmentHandler.GetCustodyChain)

		// Equipment
		r.Get("/api/equipment", equipmentHandler.GetAll)
		r.Post("/api/equipment", equipmentHandler.Create)
		r.Patch("/api/equipment/{id}/maintenance", equipmentHandler.ReportMaintenance)
		r.Patch("/api/equipment/{id}/breakdown", equipmentHandler.ReportBreakdown)
		r.Patch("/api/equipment/{id}/resolve", equipmentHandler.Resolve)

		// Dead-head pairings
		r.Get("/api/deadhead/eligible", deadheadHandler.GetEligible)
		r.Post("/api/deadhead/pair", deadheadHandler.Pair)
		r.Delete("/api/deadhead/{pairingId}", deadheadHandler.Cancel)

		// Analytics
		r.Get("/api/analytics/summary", analyticsHandler.GetSummary)
		r.Get("/api/analytics/operating-cost", analyticsHandler.GetOperatingCost)

		// Regional inventory lookup
		r.Get("/api/inventory/region", regionalInvHandler.GetRegional)

		// Invoices (read-only)
		r.Get("/api/invoice/{id}", invoiceHandler.GetInvoice)
		r.Get("/api/invoice/store/{storeId}", invoiceHandler.GetInvoicesByStore)
	})

	port := viper.GetString("PORT")
	log.Printf("starting switchyard-go on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
