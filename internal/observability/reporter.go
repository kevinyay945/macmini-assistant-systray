package observability

import (
	"context"
	"sync"
	"time"
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
	timeout   time.Duration
}

// NewMultiReporter creates a reporter that sends to multiple destinations.
func NewMultiReporter(reporters ...ErrorReporter) *MultiReporter {
	return &MultiReporter{
		reporters: reporters,
		timeout:   DefaultReportTimeout,
	}
}

// WithTimeout returns a new MultiReporter with the specified timeout.
func (m *MultiReporter) WithTimeout(d time.Duration) *MultiReporter {
	return &MultiReporter{
		reporters: m.reporters,
		timeout:   d,
	}
}

// DefaultReportTimeout is the default timeout for multi-reporter operations.
const DefaultReportTimeout = 5 * time.Second

// Report sends the error to all reporters concurrently with a timeout.
// If reporters don't complete within the timeout, the function returns anyway.
func (m *MultiReporter) Report(ctx context.Context, err error) {
	m.reportWithTimeout(ctx, func(timeoutCtx context.Context, reporter ErrorReporter) {
		reporter.Report(timeoutCtx, err)
	})
}

// ReportWithContext sends the error with context to all reporters concurrently with a timeout.
// If reporters don't complete within the timeout, the function returns anyway.
func (m *MultiReporter) ReportWithContext(ctx context.Context, err error, extra map[string]interface{}) {
	m.reportWithTimeout(ctx, func(timeoutCtx context.Context, reporter ErrorReporter) {
		reporter.ReportWithContext(timeoutCtx, err, extra)
	})
}

// reportWithTimeout executes the report function on all reporters with a timeout.
// The reportFn receives the timeout context so individual reporters can detect cancellation.
func (m *MultiReporter) reportWithTimeout(ctx context.Context, reportFn func(context.Context, ErrorReporter)) {
	timeoutCtx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	var wg sync.WaitGroup
	for _, r := range m.reporters {
		wg.Add(1)
		go func(reporter ErrorReporter) {
			defer wg.Done()
			// Check if already cancelled before starting
			select {
			case <-timeoutCtx.Done():
				return
			default:
			}
			reportFn(timeoutCtx, reporter)
		}(r)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All reporters completed
	case <-timeoutCtx.Done():
		// Timeout reached, some reporters may still be running but will see cancelled context
	}
}
