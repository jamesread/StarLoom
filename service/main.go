package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"net/http"
	"path/filepath"

	connectcors "connectrpc.com/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"

	"github.com/jamesread/starapp/service/internal/auth"
	"github.com/jamesread/starapp/service/internal/config"
	mcppkg "github.com/jamesread/starapp/service/internal/mcp"
	"github.com/jamesread/starapp/service/internal/migrate"
	"github.com/jamesread/starapp/service/internal/server"
	"github.com/jamesread/starapp/service/internal/store"
	"github.com/jamesread/starapp/service/internal/webhook"
)

//go:embed gen/openapi.json
var openAPISpec []byte

//go:embed llms.txt
var llmsTxt []byte

func main() {
	configDir := flag.String("configdir", "", "Config directory (default: ~/.config/starapp)")
	flag.Parse()

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	if err := run(log, *configDir); err != nil {
		log.WithError(err).Fatal("starapp")
	}
}

func run(log *logrus.Logger, configDir string) error {
	cfg, err := config.Load(configDir)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	log.WithFields(logrus.Fields{
		"config_dir": cfg.ConfigDir,
		"db_driver":  cfg.DBDriver,
		"db_path":    cfg.DBPath,
	}).Info("database configuration")

	st, err := openStore(cfg)
	if err != nil {
		return fmt.Errorf("database %q: open: %w", cfg.DBPath, err)
	}
	defer st.Close()

	ctx := context.Background()

	migrationsDir, err := migrate.FindSQLiteMigrationsDir()
	if err != nil {
		return fmt.Errorf("database %q: %w", cfg.DBPath, err)
	}
	log.WithFields(logrus.Fields{
		"db_path":        cfg.DBPath,
		"migrations_dir": migrationsDir,
	}).Info("running database migrations")

	applied, err := migrate.RunSQLiteUp(st, migrationsDir)
	if err != nil {
		return fmt.Errorf("database %q: migrate: %w", cfg.DBPath, err)
	}
	if applied > 0 {
		log.WithFields(logrus.Fields{
			"db_path": cfg.DBPath,
			"applied": applied,
		}).Info("database migrations applied")
	}

	if err := migrate.AssertRequired(ctx, st, config.RequiredMigration, cfg.DBPath); err != nil {
		return err
	}
	if latest, err := st.LatestMigration(ctx); err != nil {
		log.WithField("db_path", cfg.DBPath).WithError(err).Warn("could not read current database migration")
	} else {
		log.WithFields(logrus.Fields{
			"db_path":   cfg.DBPath,
			"migration": latest,
		}).Info("database schema ready")
	}

	siteTitle := "StarApp"
	showFooter := true
	if cfg != nil {
		showFooter = cfg.ShowFooter
	}
	if err := server.EnsureDefaultCvars(ctx, st, siteTitle, showFooter); err != nil {
		return fmt.Errorf("cvars: %w", err)
	}

	svc := server.New(cfg, st, &webhook.Dispatcher{Store: st}, log)
	if err := svc.BootstrapIAM(ctx); err != nil {
		return fmt.Errorf("iam bootstrap: %w", err)
	}

	authLayer := auth.NewLayer(st, cfg.Auth, log)
	apiPath, apiHandler := svc.Handler(authLayer)
	apiHandler = withCORS(apiHandler)

	mcpHandler := authLayer.WrapMCPHandler(mcppkg.NewHandler(svc))

	mux := http.NewServeMux()
	mux.Handle(apiPath, apiHandler)
	mux.Handle("/api"+apiPath, http.StripPrefix("/api", apiHandler))
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/openapi", serveOpenAPI)
	mux.HandleFunc("/llms.txt", serveLLMsTxt)
	mux.Handle("/mcp", mcpHandler)
	mux.Handle("/mcp/", mcpHandler)
	mux.Handle("/avatars/", svc.AvatarHandler(authLayer))

	if cfg.WebUIDir != "" {
		webui := filepath.Clean(cfg.WebUIDir)
		mux.Handle("/", http.FileServer(http.Dir(webui)))
		log.WithField("webui", webui).Info("serving web UI")
	}

	log.WithFields(logrus.Fields{
		"addr":      cfg.HTTPAddr,
		"db_path":   cfg.DBPath,
		"db_driver": cfg.DBDriver,
	}).Info("starapp listening")

	if err := http.ListenAndServe(cfg.HTTPAddr, mux); err != nil {
		return fmt.Errorf("http: %w", err)
	}
	return nil
}

func openStore(cfg *config.Config) (store.Store, error) {
	switch cfg.DBDriver {
	case "sqlite", "":
		return store.OpenSQLite(cfg.DBPath)
	default:
		return nil, fmt.Errorf("unsupported db_driver %q (supported: sqlite)", cfg.DBDriver)
	}
}

func withCORS(h http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
	}).Handler(h)
}

func serveOpenAPI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openAPISpec)
}

func serveLLMsTxt(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(llmsTxt)
}
