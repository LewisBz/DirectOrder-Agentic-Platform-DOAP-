package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/config"
	"github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/database"
	plathttp "github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Open(ctx, cfg.AppDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	srv := &http.Server{
		Addr:              cfg.APIAddr,
		Handler:           plathttp.NewRouter(db),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("api listening on %s", cfg.APIAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
