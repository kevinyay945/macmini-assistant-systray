package updater_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kevinyay945/macmini-assistant-systray/internal/updater"
)

func TestNewChecker(t *testing.T) {
	c := updater.NewChecker("owner/repo", "v1.0.0")

	if c == nil {
		t.Fatal("NewChecker returned nil")
	}
	if c.Repo() != "owner/repo" {
		t.Errorf("Repo() = %q, want %q", c.Repo(), "owner/repo")
	}
	if c.CurrentVersion() != "v1.0.0" {
		t.Errorf("CurrentVersion() = %q, want %q", c.CurrentVersion(), "v1.0.0")
	}
	if c.CheckInterval() != updater.DefaultCheckInterval {
		t.Errorf("CheckInterval() = %v, want %v", c.CheckInterval(), updater.DefaultCheckInterval)
	}
}

func TestChecker_WithCheckInterval(t *testing.T) {
	interval := 1 * time.Hour
	c := updater.NewChecker("owner/repo", "v1.0.0", updater.WithCheckInterval(interval))

	if c.CheckInterval() != interval {
		t.Errorf("CheckInterval() = %v, want %v", c.CheckInterval(), interval)
	}
}

func TestChecker_GetLatestRelease(t *testing.T) {
	expected := updater.Release{
		TagName:     "v2.0.0",
		Name:        "Release 2.0.0",
		Body:        "Changelog",
		PublishedAt: "2024-01-15T00:00:00Z",
		Assets: []updater.Asset{
			{
				Name:        "app_darwin_amd64.tar.gz",
				DownloadURL: "https://example.com/download",
				Size:        1024,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/releases/latest" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
			t.Errorf("unexpected Accept header: %s", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(expected); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create a checker with custom HTTP client that redirects to test server
	client := &http.Client{
		Transport: &testTransport{baseURL: server.URL},
	}
	c := updater.NewChecker("owner/repo", "v1.0.0", updater.WithHTTPClient(client))

	release, err := c.GetLatestRelease(context.Background())
	if err != nil {
		t.Fatalf("GetLatestRelease() error = %v", err)
	}

	if release.TagName != expected.TagName {
		t.Errorf("TagName = %q, want %q", release.TagName, expected.TagName)
	}
	if release.Name != expected.Name {
		t.Errorf("Name = %q, want %q", release.Name, expected.Name)
	}
	if len(release.Assets) != 1 {
		t.Errorf("len(Assets) = %d, want 1", len(release.Assets))
	}
}

func TestChecker_GetLatestRelease_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: &testTransport{baseURL: server.URL},
	}
	c := updater.NewChecker("owner/repo", "v1.0.0", updater.WithHTTPClient(client))

	_, err := c.GetLatestRelease(context.Background())
	if err == nil {
		t.Error("GetLatestRelease() expected error for 404")
	}
}

func TestChecker_GetLatestRelease_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: &testTransport{baseURL: server.URL},
	}
	c := updater.NewChecker("owner/repo", "v1.0.0", updater.WithHTTPClient(client))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := c.GetLatestRelease(ctx)
	if err == nil {
		t.Error("GetLatestRelease() expected error for canceled context")
	}
}

func TestChecker_GetLatestRelease_NetworkError(t *testing.T) {
	// Use a client that will fail to connect
	client := &http.Client{
		Transport: &testTransport{baseURL: "http://localhost:1"},
		Timeout:   100 * time.Millisecond,
	}
	c := updater.NewChecker("owner/repo", "v1.0.0", updater.WithHTTPClient(client))

	_, err := c.GetLatestRelease(context.Background())
	if err == nil {
		t.Error("GetLatestRelease() expected error for network failure")
	}
}

func TestChecker_ManualCheck_UpdateAvailable(t *testing.T) {
	release := updater.Release{
		TagName: "v2.0.0",
		Name:    "Release 2.0.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := &http.Client{
		Transport: &testTransport{baseURL: server.URL},
	}
	c := updater.NewChecker("owner/repo", "v1.0.0", updater.WithHTTPClient(client))
	u := updater.New(updater.Config{CurrentVersion: "v1.0.0"})

	result, err := c.ManualCheck(context.Background(), u)
	if err != nil {
		t.Fatalf("ManualCheck() error = %v", err)
	}
	if result == nil {
		t.Error("ManualCheck() returned nil, expected release")
	}
	if result.TagName != "v2.0.0" {
		t.Errorf("TagName = %q, want %q", result.TagName, "v2.0.0")
	}
}

func TestChecker_ManualCheck_NoUpdate(t *testing.T) {
	release := updater.Release{
		TagName: "v1.0.0",
		Name:    "Release 1.0.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := &http.Client{
		Transport: &testTransport{baseURL: server.URL},
	}
	c := updater.NewChecker("owner/repo", "v1.0.0", updater.WithHTTPClient(client))
	u := updater.New(updater.Config{CurrentVersion: "v1.0.0"})

	result, err := c.ManualCheck(context.Background(), u)
	if err != nil {
		t.Fatalf("ManualCheck() error = %v", err)
	}
	if result != nil {
		t.Errorf("ManualCheck() = %v, expected nil (no update)", result)
	}
}

func TestChecker_PeriodicCheck(t *testing.T) {
	var callCount atomic.Int32

	release := updater.Release{
		TagName: "v2.0.0",
		Name:    "Release 2.0.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := &http.Client{
		Transport: &testTransport{baseURL: server.URL},
	}
	// Use very short interval for testing
	c := updater.NewChecker("owner/repo", "v1.0.0",
		updater.WithHTTPClient(client),
		updater.WithCheckInterval(50*time.Millisecond),
	)
	u := updater.New(updater.Config{CurrentVersion: "v1.0.0"})

	var updateCount atomic.Int32
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		c.StartPeriodicCheck(ctx, u, func(_ *updater.Release) {
			updateCount.Add(1)
		})
		close(done)
	}()

	<-done

	// Should have at least the immediate check plus some periodic checks
	if callCount.Load() < 2 {
		t.Errorf("API called %d times, expected at least 2", callCount.Load())
	}
	if updateCount.Load() < 2 {
		t.Errorf("onUpdate called %d times, expected at least 2", updateCount.Load())
	}
}

func TestChecker_RateLimiting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message": "API rate limit exceeded"}`))
	}))
	defer server.Close()

	client := &http.Client{
		Transport: &testTransport{baseURL: server.URL},
	}
	c := updater.NewChecker("owner/repo", "v1.0.0", updater.WithHTTPClient(client))

	_, err := c.GetLatestRelease(context.Background())
	if err == nil {
		t.Error("GetLatestRelease() expected error for rate limiting")
	}
}

// testTransport is a custom transport that redirects GitHub API calls to a test server.
type testTransport struct {
	baseURL string
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Redirect GitHub API calls to test server
	// Parse the base URL safely to extract scheme and host
	parsedURL, err := url.Parse(t.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	req.URL.Scheme = parsedURL.Scheme
	req.URL.Host = parsedURL.Host
	return http.DefaultTransport.RoundTrip(req)
}
