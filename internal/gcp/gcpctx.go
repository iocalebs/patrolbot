package gcp

import "context"

type traceKey struct{}

func contextWithTrace(ctx context.Context, trace string) context.Context {
	return context.WithValue(ctx, traceKey{}, trace)
}

func traceFromContext(ctx context.Context) (string, bool) {
	t, ok := ctx.Value(traceKey{}).(string)

	return t, ok
}
