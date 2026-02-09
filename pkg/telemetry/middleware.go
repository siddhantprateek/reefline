package telemetry

import (
	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName = "github.com/siddhantprateek/reefline/fiber"
)

type MiddlewareConfig struct {
	TracerProvider trace.TracerProvider
	Propagators    propagation.TextMapPropagator
	SpanNameFunc   func(c fiber.Ctx) string
}

func FiberMiddleware(config ...MiddlewareConfig) fiber.Handler {
	cfg := MiddlewareConfig{}
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}

	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}

	if cfg.SpanNameFunc == nil {
		cfg.SpanNameFunc = func(c fiber.Ctx) string {
			return c.Method() + " " + c.Route().Path
		}
	}

	tracer := cfg.TracerProvider.Tracer(tracerName)

	return func(c fiber.Ctx) error {
		// Extract context from request headers
		ctx := cfg.Propagators.Extract(c.Context(), &FiberCarrier{c: c})

		spanName := cfg.SpanNameFunc(c)
		ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		// Set span attributes
		span.SetAttributes(
			semconv.HTTPMethodKey.String(c.Method()),
			semconv.HTTPURLKey.String(c.OriginalURL()),
			semconv.HTTPSchemeKey.String(c.Protocol()),
			semconv.HTTPTargetKey.String(c.Path()),
			semconv.HTTPUserAgentKey.String(string(c.Request().Header.UserAgent())),
			semconv.HTTPClientIPKey.String(c.IP()),
			attribute.String("http.route", c.Route().Path),
		)

		// Store context in fiber context
		c.Context().Value("otel-ctx")

		// Process request
		err := c.Next()

		// Set response attributes
		span.SetAttributes(
			semconv.HTTPStatusCodeKey.Int(c.Response().StatusCode()),
			attribute.String("http.response.size", string(rune(len(c.Response().Body())))),
		)

		// Set span status based on response
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		} else if c.Response().StatusCode() >= 400 {
			span.SetStatus(codes.Error, "HTTP error")
		}

		return err
	}
}

type FiberCarrier struct {
	c fiber.Ctx
}

func (fc *FiberCarrier) Get(key string) string {
	return fc.c.Get(key)
}

func (fc *FiberCarrier) Set(key, value string) {
	fc.c.Set(key, value)
}

func (fc *FiberCarrier) Keys() []string {
	return []string{}
}
