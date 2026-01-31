// Package tools provides common utilities for tool implementations.
package tools

import "fmt"

// GetRequiredString extracts a required string parameter from the params map.
// Returns an error if the parameter is missing or empty.
func GetRequiredString(params map[string]interface{}, key string) (string, error) {
	val, ok := params[key].(string)
	if !ok || val == "" {
		return "", fmt.Errorf("parameter %q is required", key)
	}
	return val, nil
}

// GetOptionalString extracts an optional string parameter with a default value.
// Returns the default value if the parameter is missing or empty.
func GetOptionalString(params map[string]interface{}, key, defaultVal string) string {
	if val, ok := params[key].(string); ok && val != "" {
		return val
	}
	return defaultVal
}

// GetOptionalInt extracts an optional int parameter with a default value.
// Returns the default value if the parameter is missing or not an int.
func GetOptionalInt(params map[string]interface{}, key string, defaultVal int) int {
	if val, ok := params[key].(int); ok {
		return val
	}
	if val, ok := params[key].(float64); ok {
		return int(val)
	}
	return defaultVal
}

// GetOptionalBool extracts an optional bool parameter with a default value.
// Returns the default value if the parameter is missing or not a bool.
func GetOptionalBool(params map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := params[key].(bool); ok {
		return val
	}
	return defaultVal
}
