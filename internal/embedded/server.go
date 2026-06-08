package embedded

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"passworder/internal/config"
	"passworder/internal/database"
	"passworder/internal/handler"
	"passworder/internal/repository"
	"passworder/internal/service"
	"passworder/internal/storage"
)

//go:embed assets/static/*
//go:embed assets/static/css/*
//go:embed assets/static/js/*
//go:embed assets/static/vendor/*
//go:embed assets/static/vendor/vditor/*
//go:embed assets/static/vendor/vditor/dist/*
//go:embed assets/static/vendor/vditor/dist/js/*
//go:embed assets/static/vendor/vditor/dist/js/i18n/*
//go:embed assets/static/vendor/vditor/dist/js/icons/*
//go:embed assets/static/vendor/vditor/dist/js/lute/*
//go:embed assets/static/vendor/vditor/dist/css/*
//go:embed assets/static/vendor/vditor/dist/css/content-theme/*
var staticFS embed.FS

//go:embed assets/migrations/*.sql
var migrationsFS embed.FS

var Version = "v1.0.2"

type EmbeddedServer struct {
	mu           sync.Mutex
	cfg          *config.Config
	db           *sqlx.DB
	httpServer   *http.Server
	listener     net.Listener
	serverErr    chan error
	started      bool
	shutdownCtx  context.Context
	shutdownFunc context.CancelFunc
}

func NewEmbeddedServer(overrides config.CLIOverrides) (*EmbeddedServer, error) {
	initialCfg := config.LoadFromEnvAndDefaults(overrides)
	db, err := openEmbeddedDatabase(initialCfg)
	if err != nil {
		return nil, err
	}

	cfg := loadRuntimeConfig(db, initialCfg, overrides)
	router, err := createEmbeddedRouter(cfg, db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	server := &EmbeddedServer{
		cfg:          cfg,
		db:           db,
		serverErr:    make(chan error, 1),
		shutdownCtx:  ctx,
		shutdownFunc: cancel,
	}
	server.httpServer = &http.Server{
		Addr:    serverAddress(cfg),
		Handler: router,
	}
	return server, nil
}

func (s *EmbeddedServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return nil
	}

	listener, err := net.Listen(listenerNetwork(s.cfg.Host), s.httpServer.Addr)
	if err != nil {
		return err
	}
	s.listener = listener
	s.started = true
	startAutoReminder(s.shutdownCtx, s.db, s.cfg.ReminderCheckInterval)

	go func() {
		err := s.httpServer.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			s.serverErr <- err
			return
		}
		s.serverErr <- nil
	}()

	return nil
}

func (s *EmbeddedServer) Wait() error {
	return <-s.serverErr
}

func (s *EmbeddedServer) Stop() error {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		if s.db != nil {
			return s.db.Close()
		}
		return nil
	}
	s.started = false
	shutdown := s.shutdownFunc
	httpServer := s.httpServer
	db := s.db
	s.mu.Unlock()

	shutdown()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
	if db != nil {
		return db.Close()
	}
	return nil
}

func (s *EmbeddedServer) Config() *config.Config {
	return s.cfg
}

func embeddedAssetsFS() fs.FS {
	assetRoot, err := fs.Sub(migrationsFS, "assets")
	if err != nil {
		panic(err)
	}
	return assetRoot
}

func openEmbeddedDatabase(cfg *config.Config) (*sqlx.DB, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		return nil, err
	}
	db, err := database.Open(cfg.DBPath)
	if err != nil {
		return nil, err
	}
	migrationRoot, err := fs.Sub(migrationsFS, "assets")
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := database.Migrate(db, migrationRoot); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func createEmbeddedRouter(cfg *config.Config, db *sqlx.DB) (http.Handler, error) {
	fileStore := storage.NewFileStorage(cfg.StorageDir)
	if err := fileStore.EnsureRoot(); err != nil {
		return nil, err
	}
	staticRoot, err := fs.Sub(staticFS, "assets")
	if err != nil {
		return nil, err
	}
	return handler.NewRouter(cfg, db, fileStore, staticRoot), nil
}

func loadRuntimeConfig(db *sqlx.DB, initialCfg *config.Config, overrides config.CLIOverrides) *config.Config {
	settingService := service.NewSettingService(repository.NewSettingRepository(db))
	return applyDatabaseConfig(initialCfg, settingService, overrides)
}

func applyDatabaseConfig(cfg *config.Config, settingService *service.SettingService, cli config.CLIOverrides) *config.Config {
	dbConfig := settingService.GetServerConfig()

	if cli.Host == "" && dbConfig.Host != "" {
		cfg.Host = dbConfig.Host
	}
	if cli.Port == 0 && dbConfig.Port > 0 {
		cfg.Port = dbConfig.Port
	}
	if cli.StorageDir == "" && dbConfig.StorageDir != "" {
		cfg.StorageDir = dbConfig.StorageDir
	}
	if cli.ReminderCheckInterval == 0 && dbConfig.ReminderCheckInterval > 0 {
		cfg.ReminderCheckInterval = dbConfig.ReminderCheckInterval
	}

	return cfg
}

func startAutoReminder(ctx context.Context, db *sqlx.DB, intervalMinutes int) {
	reminderRepo := repository.NewReminderRepository(db)
	settingService := service.NewSettingService(repository.NewSettingRepository(db))
	mailSender := service.NewSMTPSender()
	reminderService := service.NewReminderService(reminderRepo, settingService, mailSender)

	if intervalMinutes <= 0 {
		intervalMinutes = 10
	}

	go func() {
		result, err := reminderService.SendDue()
		if err == nil && result["sentGroups"] > 0 {
			fmt.Printf("Auto-reminder: startup sent %d groups, %d reminders\n", result["sentGroups"], result["sentReminders"])
		}

		ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, _ = reminderService.SendDue()
			}
		}
	}()
}

func serverAddress(cfg *config.Config) string {
	return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
}

func listenerNetwork(host string) string {
	if ip := net.ParseIP(host); ip != nil && ip.To4() != nil {
		return "tcp4"
	}
	return "tcp"
}
