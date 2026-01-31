package downloader

import (
	"testing"
)

func TestNew(t *testing.T) {
	basePath := "/tmp/downloads"
	dl := New(basePath)

	if dl == nil {
		t.Fatal("New() returned nil")
	}

	if dl.basePath != basePath {
		t.Errorf("basePath = %v, want %v", dl.basePath, basePath)
	}
}

func TestBuildDownieURL(t *testing.T) {
	dl := New("/tmp")

	tests := []struct {
		name           string
		url            string
		destination    string
		postProcessing string
		useUGE         bool
		wantContains   []string
	}{
		{
			name:           "basic URL",
			url:            "https://youtube.com/watch?v=123",
			destination:    "/tmp/test",
			postProcessing: "",
			useUGE:         false,
			wantContains:   []string{"downie://XUOpenURL?url=", "destination=/tmp/test"},
		},
		{
			name:           "URL with post processing",
			url:            "https://youtube.com/watch?v=123",
			destination:    "/tmp/test",
			postProcessing: "mp4",
			useUGE:         false,
			wantContains:   []string{"downie://XUOpenURL?url=", "postprocessing=mp4", "destination=/tmp/test"},
		},
		{
			name:           "URL with UGE",
			url:            "https://youtube.com/watch?v=123",
			destination:    "/tmp/test",
			postProcessing: "",
			useUGE:         true,
			wantContains:   []string{"downie://XUOpenURL?url=", "action=open_in_uge"},
		},
		{
			name:           "URL with special characters",
			url:            "https://youtube.com/watch?v=123&t=10",
			destination:    "/tmp/test",
			postProcessing: "",
			useUGE:         false,
			wantContains:   []string{"downie://XUOpenURL?url=", "%26t=10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dl.buildDownieURL(tt.url, tt.destination, tt.postProcessing, tt.useUGE)

			for _, want := range tt.wantContains {
				if !contains(got, want) {
					t.Errorf("buildDownieURL() = %v, want to contain %v", got, want)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
