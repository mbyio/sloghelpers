// Package slogcontexthandler provides a slog.Handler that adds attributes from
// the log's context.Context.
package slogcontexthandler

import (
	"context"
	"log/slog"
	"slices"
)

type ctxKeyType struct{}

var ctxKey = ctxKeyType{}

// GetAttrs returns the attributes set on a context.Context, if any.
func GetAttrs(ctx context.Context) []slog.Attr {
	attrs := ctx.Value(ctxKey)
	if attrs == nil {
		return nil
	}
	return attrs.([]slog.Attr)
}

// AddAttrs adds attributes to a context.Context.
//
// The attributes will be merged with any existing attributes on the context.
//
// Any new attribute with the same key as an existing attribute will replace
// the existing attribute.
func AddAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	prevAttrs := GetAttrs(ctx)
	newAttrs := make([]slog.Attr, len(prevAttrs))
	copy(newAttrs, prevAttrs)
	// TODO this might be a stupid way to deduplicate
	for _, attr := range attrs {
		i := slices.IndexFunc(newAttrs, func(newAttr slog.Attr) bool {
			return newAttr.Key == attr.Key
		})
		if i >= 0 {
			newAttrs[i] = attr
			continue
		}
		newAttrs = append(newAttrs, attr)
	}
	return SetAttrs(ctx, newAttrs...)
}

// SetAttrs sets the attributes on a context.Context.
//
// Unlike AddAttrs, SetAttrs does not deduplicate or merge the new attributes
// with existing attributes.
func SetAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	return context.WithValue(ctx, ctxKey, attrs)
}

type handlerWrapper struct {
	slog.Handler
}

func (hw *handlerWrapper) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(GetAttrs(ctx)...)
	return hw.Handler.Handle(ctx, r)
}

// New returns a slog.Handler that adds attributes from the log's context.Context.
//
// All other methods are just delegated to the wrapped slog.Handler.
func New(handler slog.Handler) slog.Handler {
	return &handlerWrapper{handler}
}
