package main

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "go.uber.org/zap"

    "roleplay/internal/config"
    "roleplay/internal/indexer"
    "roleplay/internal/repository"
    "roleplay/internal/router"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    zap.ReplaceGlobals(logger)

    // Load configuration
    if err := config.Load(); err != nil {
        zap.L().Fatal("failed to load config", zap.Error(err))
    }

    // Init MongoDB
    if err := repository.InitMongo(context.Background()); err != nil {
        zap.L().Fatal("failed to init mongo", zap.Error(err))
    }
    defer repository.CloseMongo(context.Background())

    // Ensure indexes
    if err := indexer.EnsureAllIndexes(context.Background()); err != nil {
        zap.L().Fatal("failed to ensure indexes", zap.Error(err))
    }

    r := router.New()

    srv := &http.Server{
        Addr:              fmt.Sprintf(":%d", config.C.Server.Port),
        Handler:           r,
        ReadTimeout:       15 * time.Second,
        ReadHeaderTimeout: 10 * time.Second,
        WriteTimeout:      30 * time.Second,
        IdleTimeout:       60 * time.Second,
    }

    go func() {
        zap.L().Info("server starting", zap.Int("port", config.C.Server.Port))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            zap.L().Fatal("http server error", zap.Error(err))
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        zap.L().Error("server shutdown error", zap.Error(err))
    }
    zap.L().Info("server stopped")
}

