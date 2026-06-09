package telemetry

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kostayne/go-microservice/pkg/config"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Init configures OpenTelemetry when OTEL_EXPORTER_OTLP_ENDPOINT is set.
func Init(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	endpoint := strings.TrimSpace(config.String("OTEL_EXPORTER_OTLP_ENDPOINT", ""))
	if endpoint == "" {
		log.Printf("[%s] tracing disabled (OTEL_EXPORTER_OTLP_ENDPOINT not set)", serviceName)
		return func(context.Context) error { return nil }, nil
	}

	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	samplerArg := config.Float("OTEL_TRACES_SAMPLER_ARG", 1.0)
	if samplerArg < 0 {
		samplerArg = 0
	}
	if samplerArg > 1 {
		samplerArg = 1
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("otlp exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.DeploymentEnvironment(string(config.AppEnv())),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(samplerArg))),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Printf("[%s] tracing enabled → %s (sample=%.2f)", serviceName, endpoint, samplerArg)

	return tp.Shutdown, nil
}

// Tracer returns a tracer for manual spans.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// Propagator returns the global text map propagator.
func Propagator() propagation.TextMapPropagator {
	return otel.GetTextMapPropagator()
}

// WrapHTTP instruments incoming HTTP handlers.
func WrapHTTP(serviceName string, handler http.Handler) http.Handler {
	if !Enabled() {
		return handler
	}
	return otelhttp.NewHandler(handler, serviceName)
}

// HTTPClient returns an HTTP client that propagates trace context.
func HTTPClient(timeout time.Duration) *http.Client {
	transport := http.DefaultTransport
	if Enabled() {
		transport = otelhttp.NewTransport(http.DefaultTransport)
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// Enabled reports whether OTLP export is configured.
func Enabled() bool {
	return strings.TrimSpace(config.String("OTEL_EXPORTER_OTLP_ENDPOINT", "")) != ""
}
