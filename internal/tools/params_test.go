package tools_test

import (
	"testing"

	"github.com/kevinyay945/macmini-assistant-systray/internal/tools"
)

func TestGetRequiredString_Valid(t *testing.T) {
	params := map[string]interface{}{
		"url": "https://example.com",
	}

	val, err := tools.GetRequiredString(params, "url")
	if err != nil {
		t.Errorf("GetRequiredString() returned error: %v", err)
	}
	if val != "https://example.com" {
		t.Errorf("GetRequiredString() = %q, want %q", val, "https://example.com")
	}
}

func TestGetRequiredString_Missing(t *testing.T) {
	params := map[string]interface{}{}

	_, err := tools.GetRequiredString(params, "url")
	if err == nil {
		t.Error("GetRequiredString() should return error for missing parameter")
	}
}

func TestGetRequiredString_Empty(t *testing.T) {
	params := map[string]interface{}{
		"url": "",
	}

	_, err := tools.GetRequiredString(params, "url")
	if err == nil {
		t.Error("GetRequiredString() should return error for empty parameter")
	}
}

func TestGetRequiredString_WrongType(t *testing.T) {
	params := map[string]interface{}{
		"url": 12345,
	}

	_, err := tools.GetRequiredString(params, "url")
	if err == nil {
		t.Error("GetRequiredString() should return error for wrong type")
	}
}

func TestGetOptionalString_Present(t *testing.T) {
	params := map[string]interface{}{
		"format": "mp4",
	}

	val := tools.GetOptionalString(params, "format", "avi")
	if val != "mp4" {
		t.Errorf("GetOptionalString() = %q, want %q", val, "mp4")
	}
}

func TestGetOptionalString_Missing(t *testing.T) {
	params := map[string]interface{}{}

	val := tools.GetOptionalString(params, "format", "avi")
	if val != "avi" {
		t.Errorf("GetOptionalString() = %q, want default %q", val, "avi")
	}
}

func TestGetOptionalString_Empty(t *testing.T) {
	params := map[string]interface{}{
		"format": "",
	}

	val := tools.GetOptionalString(params, "format", "avi")
	if val != "avi" {
		t.Errorf("GetOptionalString() = %q, want default %q for empty", val, "avi")
	}
}

func TestGetOptionalInt_Present(t *testing.T) {
	params := map[string]interface{}{
		"count": 10,
	}

	val := tools.GetOptionalInt(params, "count", 5)
	if val != 10 {
		t.Errorf("GetOptionalInt() = %d, want %d", val, 10)
	}
}

func TestGetOptionalInt_Float64(t *testing.T) {
	params := map[string]interface{}{
		"count": float64(10),
	}

	val := tools.GetOptionalInt(params, "count", 5)
	if val != 10 {
		t.Errorf("GetOptionalInt() = %d, want %d for float64 input", val, 10)
	}
}

func TestGetOptionalInt_Missing(t *testing.T) {
	params := map[string]interface{}{}

	val := tools.GetOptionalInt(params, "count", 5)
	if val != 5 {
		t.Errorf("GetOptionalInt() = %d, want default %d", val, 5)
	}
}

func TestGetOptionalBool_Present(t *testing.T) {
	params := map[string]interface{}{
		"enabled": true,
	}

	val := tools.GetOptionalBool(params, "enabled", false)
	if !val {
		t.Error("GetOptionalBool() = false, want true")
	}
}

func TestGetOptionalBool_Missing(t *testing.T) {
	params := map[string]interface{}{}

	val := tools.GetOptionalBool(params, "enabled", true)
	if !val {
		t.Error("GetOptionalBool() = false, want default true")
	}
}
