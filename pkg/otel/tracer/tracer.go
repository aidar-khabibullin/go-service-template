package tracer

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// tracer is a no-op by default so code (and tests) work before Init is called.
var tracer trace.Tracer = noop.NewTracerProvider().Tracer("")

func Init(t trace.Tracer) {
	tracer = t
}

func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, spanName, opts...) //nolint:spancheck // span is ended by caller
}

func SetStatus(span trace.Span, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
}
