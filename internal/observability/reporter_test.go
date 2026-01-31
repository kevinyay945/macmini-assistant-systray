package observability_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/kevinyay945/macmini-assistant-systray/internal/observability"
)

func TestNoOpReporter(t *testing.T) {
	reporter := observability.NoOpReporter{}
	ctx := context.Background()

	// Should not panic
	reporter.Report(ctx, errors.New("test error"))
	reporter.ReportWithContext(ctx, errors.New("test error"), map[string]interface{}{"key": "value"})
}

func TestLogReporter_Report(t *testing.T) {
	var buf bytes.Buffer
	logger := observability.New(
		observability.WithOutput(&buf),
		observability.WithJSON(),
	)
	reporter := observability.NewLogReporter(logger)
	ctx := context.Background()

	reporter.Report(ctx, errors.New("test error"))

	output := buf.String()
	if !strings.Contains(output, "test error") {
		t.Errorf("Report() should log the error, got: %s", output)
	}
	if !strings.Contains(output, "error reported") {
		t.Errorf("Report() should use 'error reported' message, got: %s", output)
	}
}

func TestLogReporter_ReportWithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := observability.New(
		observability.WithOutput(&buf),
		observability.WithJSON(),
	)
	reporter := observability.NewLogReporter(logger)
	ctx := context.Background()

	reporter.ReportWithContext(ctx, errors.New("test error"), map[string]interface{}{
		"user_id": "123",
		"action":  "download",
	})

	output := buf.String()
	if !strings.Contains(output, "test error") {
		t.Errorf("ReportWithContext() should log the error, got: %s", output)
	}
	if !strings.Contains(output, "123") {
		t.Errorf("ReportWithContext() should include extra context, got: %s", output)
	}
}

func TestLogReporter_WithRequestIDInContext(t *testing.T) {
	var buf bytes.Buffer
	logger := observability.New(
		observability.WithOutput(&buf),
		observability.WithJSON(),
	)
	reporter := observability.NewLogReporter(logger)
	ctx := observability.ContextWithRequestID(context.Background(), "req-456")

	reporter.Report(ctx, errors.New("test error"))

	output := buf.String()
	if !strings.Contains(output, "req-456") {
		t.Errorf("Report() should include request_id from context, got: %s", output)
	}
}

func TestLogReporter_WithAppError(t *testing.T) {
	var buf bytes.Buffer
	logger := observability.New(
		observability.WithOutput(&buf),
		observability.WithJSON(),
	)
	reporter := observability.NewLogReporter(logger)
	ctx := context.Background()

	appErr := observability.NewAppError("TEST_CODE", "test message").
		WithRequestID("req-789").
		WithExtra("tool", "youtube_download")

	reporter.Report(ctx, appErr)

	output := buf.String()
	if !strings.Contains(output, "TEST_CODE") {
		t.Errorf("Report() should include error_code for AppError, got: %s", output)
	}
	if !strings.Contains(output, "req-789") {
		t.Errorf("Report() should include request_id from AppError, got: %s", output)
	}
	if !strings.Contains(output, "youtube_download") {
		t.Errorf("Report() should include extra fields from AppError, got: %s", output)
	}
}

func TestMultiReporter(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	logger1 := observability.New(observability.WithOutput(&buf1), observability.WithJSON())
	logger2 := observability.New(observability.WithOutput(&buf2), observability.WithJSON())

	reporter := observability.NewMultiReporter(
		observability.NewLogReporter(logger1),
		observability.NewLogReporter(logger2),
	)
	ctx := context.Background()

	reporter.Report(ctx, errors.New("multi test error"))

	if !strings.Contains(buf1.String(), "multi test error") {
		t.Error("MultiReporter should send to first reporter")
	}
	if !strings.Contains(buf2.String(), "multi test error") {
		t.Error("MultiReporter should send to second reporter")
	}
}

func TestMultiReporter_ReportWithContext(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	logger1 := observability.New(observability.WithOutput(&buf1), observability.WithJSON())
	logger2 := observability.New(observability.WithOutput(&buf2), observability.WithJSON())

	reporter := observability.NewMultiReporter(
		observability.NewLogReporter(logger1),
		observability.NewLogReporter(logger2),
	)
	ctx := context.Background()

	reporter.ReportWithContext(ctx, errors.New("multi test"), map[string]interface{}{"key": "value"})

	if !strings.Contains(buf1.String(), "key") {
		t.Error("MultiReporter should include context in first reporter")
	}
	if !strings.Contains(buf2.String(), "key") {
		t.Error("MultiReporter should include context in second reporter")
	}
}

func TestMultiReporter_Timeout(t *testing.T) {
	// Test that MultiReporter doesn't block forever if a reporter is slow
	// Use a channel to track when the fast reporter completes
	fastReporterDone := make(chan struct{})
	fastReporter := &trackingTestReporter{doneCh: fastReporterDone}

	// Create a slow reporter that takes longer than the timeout
	slowReporter := &slowTestReporter{delay: 10 * time.Second}

	reporter := observability.NewMultiReporter(
		fastReporter,
		slowReporter,
	)
	ctx := context.Background()

	start := time.Now()
	reporter.Report(ctx, errors.New("timeout test"))
	elapsed := time.Since(start)

	// Should complete within the default timeout (5s) + some buffer
	if elapsed > 6*time.Second {
		t.Errorf("MultiReporter took too long: %v (expected < 6s)", elapsed)
	}

	// Fast reporter should have completed
	select {
	case <-fastReporterDone:
		// Good, fast reporter completed
	default:
		t.Error("Fast reporter should have completed")
	}
}

// trackingTestReporter tracks when Report is called.
type trackingTestReporter struct {
	doneCh chan struct{}
}

func (tr *trackingTestReporter) Report(_ context.Context, _ error) {
	close(tr.doneCh)
}

func (tr *trackingTestReporter) ReportWithContext(_ context.Context, _ error, _ map[string]interface{}) {
	close(tr.doneCh)
}

// slowTestReporter is a test reporter that deliberately takes a long time.
type slowTestReporter struct {
	delay time.Duration
}

func (s *slowTestReporter) Report(ctx context.Context, _ error) {
	select {
	case <-time.After(s.delay):
	case <-ctx.Done():
	}
}

func (s *slowTestReporter) ReportWithContext(ctx context.Context, err error, _ map[string]interface{}) {
	s.Report(ctx, err)
}

func TestRequestID_Propagation(t *testing.T) {
	// Test that request ID flows through the entire chain
	requestID := "test-request-id-xyz"
	ctx := observability.ContextWithRequestID(context.Background(), requestID)

	// Verify it's retrievable
	got := observability.RequestIDFromContext(ctx)
	if got != requestID {
		t.Errorf("RequestIDFromContext() = %q, want %q", got, requestID)
	}

	// Verify it appears in logs
	var buf bytes.Buffer
	logger := observability.New(observability.WithOutput(&buf), observability.WithJSON())
	reporter := observability.NewLogReporter(logger)

	reporter.Report(ctx, errors.New("test"))

	if !strings.Contains(buf.String(), requestID) {
		t.Errorf("Request ID should appear in log output, got: %s", buf.String())
	}
}
