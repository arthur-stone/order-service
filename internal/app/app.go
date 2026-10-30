package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arthur-stone/order-service/internal/config"
	"github.com/arthur-stone/order-service/internal/handler"
	"github.com/arthur-stone/order-service/internal/repository"
	"github.com/arthur-stone/order-service/internal/service"
)

type App struct {
	cfg       *config.Config
	logger    *slog.Logger
	server    *http.Server
	buildInfo handler.BuildInfo
}

func New(cfg *config.Config, logger *slog.Logger, buildInfo handler.BuildInfo) *App {
	repo := repository.NewMemoryOrderRepository()
	orderSvc := service.NewOrderService(repo)
	orderHdr := handler.NewOrderHandler(orderSvc)
	healthHdr := handler.NewHealthHandler(buildInfo)

	mux := http.NewServeMux()

	// Probes & Meta
	mux.HandleFunc("GET /healthz", healthHdr.Healthz)
	mux.HandleFunc("GET /ready", healthHdr.Ready)
	mux.HandleFunc("GET /version", healthHdr.Version)

	// Orders API
	mux.HandleFunc("POST /api/v1/orders", orderHdr.Create)
	mux.HandleFunc("GET /api/v1/orders/{id}", orderHdr.GetByID)
	mux.HandleFunc("GET /api/v1/orders", orderHdr.List)
	mux.HandleFunc("POST /api/v1/orders/{id}/pay", orderHdr.Pay)
	mux.HandleFunc("POST /api/v1/orders/{id}/cancel", orderHdr.Cancel)

	loggedMux := loggingMiddleware(logger, mux)

	srv := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      loggedMux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return &App{
		cfg:       cfg,
		logger:    logger,
		server:    srv,
		buildInfo: buildInfo,
	}
}

func (a *App) Run() error {
	shutdownErr := make(chan error, 1)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigChan

		a.logger.Info("shutting down server...", slog.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
		defer cancel()

		shutdownErr <- a.server.Shutdown(ctx)
	}()

	a.logger.Info("server is starting",
		slog.String("addr", a.cfg.Addr()),
		slog.String("env", a.cfg.Env),
		slog.String("version", a.buildInfo.Version),
	)

	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server ListenAndServe: %w", err)
	}

	if err := <-shutdownErr; err != nil {
		return fmt.Errorf("server Shutdown: %w", err)
	}

	a.logger.Info("server gracefully stopped")
	return nil
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		logger.Info("http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rw.status),
			slog.Duration("duration", time.Since(start)),
			slog.String("remote_addr", r.RemoteAddr),
		)
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *statusResponseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
