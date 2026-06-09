package kafka

import (
	"context"

	"github.com/kostayne/go-microservice/pkg/telemetry"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func injectTraceContext(ctx context.Context) []kafka.Header {
	if !telemetry.Enabled() {
		return nil
	}
	carrier := propagation.MapCarrier{}
	telemetry.Propagator().Inject(ctx, carrier)
	headers := make([]kafka.Header, 0, len(carrier))
	for k, v := range carrier {
		headers = append(headers, kafka.Header{Key: k, Value: []byte(v)})
	}
	return headers
}

func extractTraceContext(headers []kafka.Header) context.Context {
	if !telemetry.Enabled() || len(headers) == 0 {
		return context.Background()
	}
	carrier := propagation.MapCarrier{}
	for _, h := range headers {
		carrier[h.Key] = string(h.Value)
	}
	return telemetry.Propagator().Extract(context.Background(), carrier)
}

func startConsumerSpan(ctx context.Context, topic string) (context.Context, trace.Span) {
	if !telemetry.Enabled() {
		return ctx, trace.SpanFromContext(ctx)
	}
	tracer := telemetry.Tracer("kafka.consumer")
	return tracer.Start(ctx, "kafka.consume "+topic,
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination", topic),
		),
	)
}

func startProducerSpan(ctx context.Context, topic, eventType string) (context.Context, trace.Span) {
	if !telemetry.Enabled() {
		return ctx, trace.SpanFromContext(ctx)
	}
	tracer := telemetry.Tracer("kafka.producer")
	return tracer.Start(ctx, "kafka.publish "+topic,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination", topic),
			attribute.String("messaging.event_type", eventType),
		),
	)
}

func recordSpanError(span trace.Span, err error) {
	if span == nil || !span.IsRecording() || err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
