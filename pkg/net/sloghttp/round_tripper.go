package sloghttp

import (
	"log/slog"
	"net/http"
	"time"
)

type slogRoundTripper struct {
	logger *slog.Logger
	rt     http.RoundTripper
}

func (srt *slogRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	startAt := time.Now()
	defer func() {
		endAt := time.Now()
		var statusCode int
		if resp != nil {
			statusCode = resp.StatusCode
		}
		// TODO record request and response body if available
		// TODO consider reporting at TRACE or DEBUG level
		srt.logger.DebugContext(
			req.Context(),
			"outbound HTTP request",
			"method", req.Method,
			"url", req.URL.String(),
			"status_code", statusCode,
			"duration", endAt.Sub(startAt),
			"error", err,
		)
	}()
	return srt.rt.RoundTrip(req)
}

// NewRoundTripper returns a new http.RoundTripper that wraps another RoundTripper and logs
// requests and responses.
func NewRoundTripper(logger *slog.Logger, rt http.RoundTripper) http.RoundTripper {
	return &slogRoundTripper{
		logger: logger,
		rt:     rt,
	}
}
