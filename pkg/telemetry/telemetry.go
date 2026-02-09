package telemetry

import (
	"context"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type TelemetryConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Enabled        bool
}

func Initialize(config TelemetryConfig) func() {
	if !config.Enabled {
		log.Println("Telemetry is disabled")
		return func() {}
	}

	log.Printf("Initializing telemetry for service: %s", config.ServiceName)

	// Create resource with service information
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String(config.ServiceVersion),
			attribute.String("environment", config.Environment),
		),
	)
	if err != nil {
		log.Printf("Failed to create resource: %v", err)
		return func() {}
	}

	// Create stdout exporter for traces
	exporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		log.Printf("Failed to create trace exporter: %v", err)
		return func() {}
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Println("Telemetry initialized successfully")

	// Return shutdown function
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
		log.Println("Telemetry shutdown complete")
	}
}

func GetConfigFromEnv() TelemetryConfig {
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "reefline-server"
	}

	serviceVersion := os.Getenv("OTEL_SERVICE_VERSION")
	if serviceVersion == "" {
		serviceVersion = "1.0.0"
	}

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	enabled := os.Getenv("OTEL_ENABLED")
	isEnabled := enabled == "true" || enabled == "1"

	return TelemetryConfig{
		ServiceName:    serviceName,
		ServiceVersion: serviceVersion,
		Environment:    environment,
		Enabled:        isEnabled,
	}
}

func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
