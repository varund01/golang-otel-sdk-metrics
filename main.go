package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var (
	tracer = otel.Tracer("sample-otel-app")
	meter  = otel.Meter("sample-otel-app")
)

func main() {
	ctx := context.Background()

	// Initialize telemetry
	cleanupTracer, err := initTracing(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize tracing: %v", err)
	}
	defer cleanupTracer(ctx)

	cleanupMetrics, err := initMetrics(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize metrics: %v", err)
	}
	defer cleanupMetrics(ctx)

	// Create an HTTP server with instrumentation
	handler := http.HandlerFunc(handleRequest)
	instrumentedHandler := otelhttp.NewHandler(handler, "http-server")

	server := &http.Server{
		Addr:    ":8080",
		Handler: instrumentedHandler,
	}

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		<-sigCh
		log.Println("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
	}()

	log.Println("Server starting on :8080")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Create a new span for this request
	ctx, span := tracer.Start(ctx, "handle-request")
	defer span.End()

	// Add attributes to the span
	span.SetAttributes(
		attribute.String("endpoint", r.URL.Path),
		attribute.String("method", r.Method),
	)

	// Record request metrics
	counter, _ := meter.Int64Counter("request_counter")
	counter.Add(ctx, 1, metric.WithAttributes(attribute.String("endpoint", r.URL.Path)))

	// Simulate some work
	time.Sleep(50 * time.Millisecond)
	
	// Log the request
	log.Printf("[%s] Processing request: %s %s", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)

	// Build the response
	response := map[string]string{"status": "blue with green"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func initTracing(ctx context.Context) (func(context.Context) error, error) {
	// Create a resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("sample-otel-app"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Configure OTLP exporter
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:4318" // Default to local collector
	}

	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Create trace provider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set the global trace provider
	otel.SetTracerProvider(tracerProvider)

	// Return cleanup function
	return tracerProvider.Shutdown, nil
}

func initMetrics(ctx context.Context) (func(context.Context) error, error) {
	// Create a resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("sample-otel-app"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create a Prometheus registry
	registry := prometheus.NewRegistry()
	
	// Create an OpenTelemetry Prometheus exporter
	exporter, err := otelprometheus.New(otelprometheus.WithRegisterer(registry))
	if err != nil {
		return nil, err
	}

	// Create meter provider
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(res),
	)

	// Set the global meter provider
	otel.SetMeterProvider(meterProvider)

	// Create a handler that exposes Prometheus metrics
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	// Create a dedicated HTTP server for metrics
	mux := http.NewServeMux()
	mux.Handle("/metrics", handler)
	
	go func() {
		log.Println("Metrics exposed on :2222/metrics")
		if err := http.ListenAndServe(":2222", mux); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// Return cleanup function
	return meterProvider.Shutdown, nil
}