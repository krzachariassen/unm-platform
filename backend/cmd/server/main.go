package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/krzachariassen/unm-platform/internal/adapter/handler"
	"github.com/krzachariassen/unm-platform/internal/adapter/repository"
	"github.com/krzachariassen/unm-platform/internal/domain/service"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/ai"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/config"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/persistence"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

func main() {
	cfg, err := config.LoadConfig(os.Getenv("UNM_ENV"))
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// AI client is optional — nil when not enabled or API key not set.
	var aiClient *ai.OpenAIClient
	if cfg.AI.Enabled && cfg.AI.APIKey != "" {
		if c, err := ai.NewOpenAIClientFromConfig(cfg.AI); err == nil {
			aiClient = c
			log.Println("AI advisor enabled")
		} else {
			log.Printf("AI advisor init failed: %v", err)
		}
	} else {
		log.Println("AI advisor disabled")
	}

	var store usecase.ModelRepository
	var csStore usecase.ChangesetRepository

	switch cfg.Storage.Driver {
	case "postgres":
		dbURL := cfg.Storage.DatabaseURL
		if dbURL == "" {
			log.Fatal("storage.database_url must be set when driver=postgres")
		}
		if cfg.Storage.MigrateOnStartup {
			if err := persistence.RunMigrations(dbURL); err != nil {
				log.Fatalf("migrations failed: %v", err)
			}
			log.Println("database migrations applied")
		}
		poolCfg, err := pgxpool.ParseConfig(dbURL)
		if err != nil {
			log.Fatalf("parse database URL: %v", err)
		}
		if cfg.Storage.MaxConnections > 0 {
			poolCfg.MaxConns = int32(cfg.Storage.MaxConnections)
		}
		db, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
		if err != nil {
			log.Fatalf("open postgres pool: %v", err)
		}
		pgModel, err := persistence.NewPGModelStore(db)
		if err != nil {
			log.Fatalf("init PGModelStore: %v", err)
		}
		pgCS := persistence.NewPGChangesetStore(db, pgModel.SystemUserID())
		pgModel.SetOnDelete(func(modelID string) {
			if n := pgCS.DeleteForModel(modelID); n > 0 {
				log.Printf("cascade-deleted %d changesets for model %s", n, modelID)
			}
		})
		store = pgModel
		csStore = pgCS
		log.Printf("storage: postgres (%s)", dbURL)
		if cfg.Server.SessionTTL > 0 {
			log.Println("session TTL eviction is memory-only; ignored for postgres driver")
		}
	default:
		memStore := repository.NewModelStore()
		memCS := repository.NewChangesetStore()
		memStore.SetOnDelete(func(modelID string) {
			if n := memCS.DeleteForModel(modelID); n > 0 {
				log.Printf("cascade-deleted %d changesets for model %s", n, modelID)
			}
		})
		if cfg.Server.SessionTTL > 0 {
			memStore.StartEviction(cfg.Server.SessionTTL, 5*time.Minute)
			defer memStore.StopEviction()
		}
		store = memStore
		csStore = memCS
		log.Println("storage: memory")
	}
	h := handler.New(handler.HandlerDeps{
		Config:            *cfg,
		ParseAndValidate:  usecase.NewParseAndValidate(parser.NewYAMLParser(), service.NewValidationEngine()),
		Fragmentation:     analyzer.NewFragmentationAnalyzer(),
		CognitiveLoad:     analyzer.NewCognitiveLoadAnalyzer(cfg.Analysis.CognitiveLoad, cfg.Analysis.InteractionWeights),
		Dependency:        analyzer.NewDependencyAnalyzer(),
		Gap:               analyzer.NewGapAnalyzer(),
		Bottleneck:        analyzer.NewBottleneckAnalyzer(cfg.Analysis.Bottleneck),
		Coupling:          analyzer.NewCouplingAnalyzer(),
		Complexity:        analyzer.NewComplexityAnalyzer(),
		Interactions:      analyzer.NewInteractionDiversityAnalyzer(cfg.Analysis.Signals),
		Unlinked:          analyzer.NewUnlinkedCapabilityAnalyzer(),
		SignalSuggestions: analyzer.NewSignalSuggestionGenerator(cfg.Analysis.Signals),
		ValueChain: analyzer.NewValueChainAnalyzerWithCogLoad(
			cfg.Analysis.ValueChain,
			analyzer.NewCognitiveLoadAnalyzer(cfg.Analysis.CognitiveLoad, cfg.Analysis.InteractionWeights),
		),
		ValueStream:    analyzer.NewValueStreamAnalyzer(),
		ChangesetStore: csStore,
		ImpactAnalyzer: analyzer.NewImpactAnalyzer(cfg.Analysis),
		AIClient:       aiClient,
		Store:          store,
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	writeTimeout := cfg.Server.WriteTimeout
	if cfg.AI.Enabled && writeTimeout < 310*time.Second {
		writeTimeout = 310 * time.Second
	}
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler.NewRouter(h, *cfg),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: writeTimeout,
	}

	go func() {
		log.Printf("server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown: %v", err)
	}
	log.Println("server stopped")
}
