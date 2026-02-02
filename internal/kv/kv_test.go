package kv

import (
	"testing"
)

func TestNormalizeNamespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string returns default",
			input:    "",
			expected: "default",
		},
		{
			name:     "non-empty string returns itself",
			input:    "myapp",
			expected: "myapp",
		},
		{
			name:     "spaces preserved",
			input:    "my app",
			expected: "my app",
		},
		{
			name:     "special characters preserved",
			input:    "app-123_test",
			expected: "app-123_test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeNamespace(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeNamespace(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDefaultNamespaceConstant(t *testing.T) {
	if DefaultNamespace != "default" {
		t.Errorf("DefaultNamespace constant = %q, want %q", DefaultNamespace, "default")
	}
}

func TestErrors(t *testing.T) {
	// Test that error constants are defined
	if ErrKeyNotFound == nil {
		t.Error("ErrKeyNotFound should not be nil")
	}

	if ErrConnectionFailed == nil {
		t.Error("ErrConnectionFailed should not be nil")
	}

	// Test error messages
	if ErrKeyNotFound.Error() != "key not found" {
		t.Errorf("ErrKeyNotFound message = %q, want %q", ErrKeyNotFound.Error(), "key not found")
	}

	if ErrConnectionFailed.Error() != "connection failed" {
		t.Errorf("ErrConnectionFailed message = %q, want %q", ErrConnectionFailed.Error(), "connection failed")
	}
}
