package config

import (
	"context"
	"errors"
)

type configContextKey struct{}

// ErrNotInContext indicates that [FromContext] was called before a context was added with [NewContext].
var ErrNotInContext = errors.New("no config in context")

// NewContext returns a derived context with the given [Config] value.
func NewContext(ctx context.Context, cfg Config) context.Context {
	return context.WithValue(ctx, configContextKey{}, cfg)
}

// FromContext returns the [Config] stored in the given context, or [ErrNotInContext] if none was found.
func FromContext(ctx context.Context) (Config, error) {
	cfg, ok := ctx.Value(configContextKey{}).(Config)
	if !ok {
		return Config{}, ErrNotInContext
	}

	return cfg, nil
}
