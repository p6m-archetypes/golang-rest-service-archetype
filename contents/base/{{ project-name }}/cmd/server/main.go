package main

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

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"{{ module_path }}/internal/config"
	"{{ module_path }}/internal/handler"
	"{{ module_path }}/internal/repository"
	"{{ module_path }}/internal/telemetry"
{% if has_persistence %}	"{{ module_path }}/internal/persistence"
{% endif %}{% if has_cache %}	"{{ module_path }}/internal/cache"
{% endif %}{% if has_messaging %}	"{{ module_path }}/internal/messaging"
{% endif %}{% if has_s3 or has_azure_blob %}	"{{ module_path }}/internal/storage"
{% endif %})

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	// Structured logging: JSON in production, text in development
	var logHandler slog.Handler
	if cfg.LoggingJSON {
		logHandler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, nil)
	}
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	// Traces: fail-open OTLP — exports iff OTEL_EXPORTER_OTLP_ENDPOINT is set
	otelShutdown := telemetry.Init(context.Background(), "{{ project-name }}")
	defer otelShutdown(context.Background())

{% if persistence == "PostgreSQL" %}	if err := persistence.Init(cfg.DatabaseURL); err != nil {
		slog.Error("persistence init failed", "error", err)
		os.Exit(1)
	}
	defer persistence.Close()
{% elseif persistence == "MySQL" %}	if err := persistence.Init(cfg.DatabaseURL); err != nil {
		slog.Error("persistence init failed", "error", err)
		os.Exit(1)
	}
	defer persistence.Close()
{% endif %}{% if has_cache %}	if err := cache.Init(cfg.RedisURL); err != nil {
		slog.Error("cache init failed", "error", err)
		os.Exit(1)
	}
	defer cache.Close()
{% endif %}{% if messaging == "Kafka" %}	if err := messaging.Init(cfg.KafkaBrokers, cfg.KafkaUsername, cfg.KafkaPassword, cfg.KafkaSASLMechanism); err != nil {
		slog.Error("messaging init failed", "error", err)
		os.Exit(1)
	}
	defer messaging.Close()
{% elseif messaging == "Pulsar" %}	if err := messaging.Init(cfg.PulsarBrokerURL, cfg.PulsarTopic, cfg.PulsarJWTToken, cfg.PulsarSubscriptionName); err != nil {
		slog.Error("messaging init failed", "error", err)
		os.Exit(1)
	}
	defer messaging.Close()
{% endif %}{% if has_s3 %}	if err := storage.InitS3(cfg.S3); err != nil {
		slog.Error("storage s3 init failed", "error", err)
		os.Exit(1)
	}
{% endif %}{% if has_azure_blob %}	if err := storage.InitAzureBlob(cfg.Azure); err != nil {
		slog.Error("storage azure-blob init failed", "error", err)
		os.Exit(1)
	}
{% endif %}
	store, err := repository.New(context.Background())
	if err != nil {
		slog.Error("repository init failed", "error", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: handler.New(store),
	}

	mgmtMux := http.NewServeMux()
	mgmtMux.HandleFunc("/health/readiness", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
{% if persistence == "PostgreSQL" %}		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := persistence.DB().Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintln(w, `{"status":"unavailable"}`)
			return
		}
{% elseif persistence == "MySQL" %}		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := persistence.DB().PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintln(w, `{"status":"unavailable"}`)
			return
		}
{% endif %}		fmt.Fprintln(w, `{"status":"ok"}`)
	})
	mgmtMux.HandleFunc("/health/liveness", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok"}`)
	})
	mgmtMux.Handle("/metrics", promhttp.Handler())
	mgmt := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.ManagementPort),
		Handler: mgmtMux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("management server starting", "port", cfg.ManagementPort)
		if err := mgmt.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("management server error", "error", err)
		}
	}()

	go func() {
		slog.Info("service server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("service server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down...")

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutCtx)
	_ = mgmt.Shutdown(shutCtx)
}
