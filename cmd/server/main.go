package main

import (
	"context"
	"gophkeep/internal/config"
	"gophkeep/internal/database"
	"gophkeep/internal/handler"
	"gophkeep/internal/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func main() {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	sugar := *zapLogger.Sugar()
	logger.Sugar = sugar

	cfg := config.MakeConfig()

	ctx := context.Background()

	env := &handler.Env{
		ConfigStruct: cfg,
		Storage:      database.NewDB(ctx, cfg.FlagDBConnectionAddress),
	}

	defer env.Storage.Close()

	r := chi.NewRouter()
	r.Use(logger.LoggingMiddleware)

	r.Get(`/ping`, env.PingDBHandle)
	r.Get(`/api/user/sync`, env.SyncHandle)
	r.Get("/api/read", env.ReadHandle)
	r.Get("/api/readfile", env.ReadFileHandle)

	r.Post("/api/user/register", env.RegisterHandle)
	r.Post("/api/user/login", env.AuthHandle)
	r.Post("/api/keepfile", env.KeepFileHandle)
	r.Post("/api/keep", env.KeepHandle)
	r.Post("/api/delete", env.DeleteHandle)
	r.Post("/api/edit", env.EditHandle)

	sugar.Infow(
		"Starting server",
		"addr", cfg.FlagRunAddr,
	)

	server := &http.Server{
		Addr:    cfg.FlagRunAddr,
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		shutdownRelease()
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	err = zapLogger.Sync()
	if err != nil {
		log.Fatal(err)
	}

	env.Storage.Close()
	sugar.Infow("Server stopped")
}
