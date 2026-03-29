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

	"github.com/krzachariassen/unm-platform/internal/adapter/handler"
	"github.com/krzachariassen/unm-platform/internal/adapter/repository"
	"github.com/krzachariassen/unm-platform/internal/domain/service"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/ai"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/config"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser"
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

	store := repository.NewModelStore()
	csStore := repository.NewChangesetStore()
	store.SetOnDelete(func(modelID string) {
		if n := csStore.DeleteForModel(modelID); n > 0 {
			log.Printf("cascade-deleted %d changesets for model %s", n, modelID)
		}
	})
	if cfg.Server.SessionTTL > 0 {
		store.StartEviction(cfg.Server.SessionTTL, 5*time.Minute)
		defer store.StopEviction()
	}
	h := handler.New(
		*cfg,
		usecase.NewParseAndValidate(parser.NewYAMLParser(), service.NewValidationEngine()),
		analyzer.NewFragmentationAnalyzer(),
		analyzer.NewCognitiveLoadAnalyzer(cfg.Analysis.CognitiveLoad, cfg.Analysis.InteractionWeights),
		analyzer.NewDependencyAnalyzer(),
		analyzer.NewGapAnalyzer(),
		analyzer.NewBottleneckAnalyzer(cfg.Analysis.Bottleneck),
		analyzer.NewCouplingAnalyzer(),
		analyzer.NewComplexityAnalyzer(),
		analyzer.NewInteractionDiversityAnalyzer(cfg.Analysis.Signals),
		analyzer.NewUnlinkedCapabilityAnalyzer(),
		analyzer.NewSignalSuggestionGenerator(cfg.Analysis.Signals),
		analyzer.NewValueChainAnalyzerWithCogLoad(
			cfg.Analysis.ValueChain,
			analyzer.NewCognitiveLoadAnalyzer(cfg.Analysis.CognitiveLoad, cfg.Analysis.InteractionWeights),
		),
		analyzer.NewValueStreamAnalyzer(),
		csStore,
		analyzer.NewImpactAnalyzer(cfg.Analysis),
		aiClient,
		store,
	)

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
