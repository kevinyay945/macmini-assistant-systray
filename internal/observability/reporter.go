package observability

import (
	"context"
	"sync"
)

// ErrorReporter is an interface for reporting errors to external systems.
type ErrorReporter interface {
	// Report sends an error to the reporting system.
	Report(ctx context.Context, err error)

	// ReportWithContext sends an error with additional context.
	ReportWithContext(ctx context.Context, err error, extra map[string]interface{})
}

// NoOpReporter is an ErrorReporter that does nothing.
// Useful for testing and as a default when no reporter is configured.
type NoOpReporter struct{}

// Report implements ErrorReporter.
func (NoOpReporter) Report(_ context.Context, _ error) {}

// ReportWithContext implements ErrorReporter.
func (NoOpReporter) ReportWithContext(_ context.Context, _ error, _ map[string]interface{}) {}

// LogReporter reports errors by logging them.
type LogReporter struct {
	logger *Logger
}

// NewLogReporter creates a new LogReporter.
func NewLogReporter(logger *Logger) *LogReporter {
	return &LogReporter{logger: logger}
}

// Report logs the error.
func (r *LogReporter) Report(ctx context.Context, err error) {
	r.ReportWithContext(ctx, err, nil)
}

// ReportWithContext logs the error with additional context.
// Request ID priority: AppError.RequestID > context request ID
func (r *LogReporter) ReportWithContext(ctx context.Context, err error, extra map[string]interface{}) {
	attrs := make([]any, 0, len(extra)*2+6)
	attrs = append(attrs, "error", err.Error())

	// Determine request ID: AppError takes priority over context
	requestID := RequestIDFromContext(ctx)

	// Add AppError specific fields
	if appErr, ok := GetAppError(err); ok {
		attrs = append(attrs, "error_code", appErr.Code)
		if appErr.RequestID != "" {
			requestID = appErr.RequestID // AppError's RequestID takes priority
		}
		for k, v := range appErr.Extra {
			attrs = append(attrs, k, v)
		}
	}

	// Add request ID once (after determining priority)
	if requestID != "" {
		attrs = append(attrs, "request_id", requestID)
	}

	// Add extra context
	for k, v := range extra {
		attrs = append(attrs, k, v)
	}

	r.logger.Error(ctx, "error reported", attrs...)
}

// MultiReporter sends errors to multiple reporters.
type MultiReporter struct {
	reporters []ErrorReporter
}

// NewMultiReporter creates a reporter that sends to multiple destinations.
func NewMultiReporter(reporters ...ErrorReporter) *MultiReporter {
	return &MultiReporter{reporters: reporters}
}

// Report sends the error to all reporters concurrently.
func (m *MultiReporter) Report(ctx context.Context, err error) {
	var wg sync.WaitGroup
	for _, r := range m.reporters {
		wg.Add(1)
		go func(reporter ErrorReporter) {
			defer wg.Done()
			reporter.Report(ctx, err)
		}(r)
	}
	wg.Wait()
}

// ReportWithContext sends the error with context to all reporters concurrently.
func (m *MultiReporter) ReportWithContext(ctx context.Context, err error, extra map[string]interface{}) {
	var wg sync.WaitGroup
	for _, r := range m.reporters {
		wg.Add(1)
		go func(reporter ErrorReporter) {
			defer wg.Done()
			reporter.ReportWithContext(ctx, err, extra)
		}(r)
	}
	wg.Wait()
}
