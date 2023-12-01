package slogpgxv5querytracer

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type SlogPgxQueryTracer struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *SlogPgxQueryTracer {
	return &SlogPgxQueryTracer{
		logger: logger,
	}
}

type startDataKeyType struct{}

var startDataKey startDataKeyType

type startData struct {
	startAt time.Time
	sql     string
}

func (s *SlogPgxQueryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	// Record data in the context. We'll log it when the query ends.
	return context.WithValue(ctx, startDataKey, &startData{
		startAt: time.Now(),
		sql:     data.SQL,
	})
}

func (s *SlogPgxQueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	sd := ctx.Value(startDataKey).(*startData)
	endAt := time.Now()
	s.logger.DebugContext(ctx, "sql query",
		"query", sd.sql,
		"startAt", sd.startAt,
		"endAt", endAt,
		"duration", endAt.Sub(sd.startAt),
		"error", data.Err,
	)
}
