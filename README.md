This module contains helpers for the `log/slog` package. https://pkg.go.dev/log/slog

Right now I am just copy-and-pasting helpers as I write/use them in my own
projects. Please let me know if you are using this module yourself or if you
have any suggestions for improvements.

# Features per third party module/package

## `context` (github.com/mbyio/sloghelpers/pkg/slogcontext)

  - a `slog.Handler` that wraps another `slog.Handler` and adds attributes from
    the context.Context to the log record.

## `net/http` (github.com/mbyio/sloghelpers/pkg/net/sloghttp)

  - a `http.RoundTripper` that emits a log for each outbound HTTP request

  - a HTTP middleware that logs information about each inbound HTTP request

## `github.com/jackc/pgx/v5` (github.com/mbyio/sloghelpers/pkg/github.com/jackc/pgx/v5/slogpgxv5)

  - a pgx query tracer that logs every SQL query, including the query text, duration, and error (if any)

    - query parameters are not logged for security reasons

## `google.golang.org/grpc` (github.com/mbyio/sloghelpers/pkg/google.golang.org/sloggrpc

  - dial options to add logging to outbound GRPC requests

  - you can use this to add logging to GCP client libraries
