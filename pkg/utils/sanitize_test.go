package utils

import (
	"strings"
	"testing"
)

func TestSanitizeLabelValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "unknown",
		},
		{
			name:     "normal alphanumeric",
			input:    "test123",
			expected: "test123",
		},
		{
			name:     "with special characters",
			input:    "test<script>alert('xss')</script>",
			expected: "test_script_alert__xss___/script_",
		},
		{
			name:     "with sql injection attempt",
			input:    "'; DROP TABLE users--",
			expected: "__ DROP TABLE users--",
		},
		{
			name:     "with newlines and tabs",
			input:    "test\n\tvalue",
			expected: "test__value",
		},
		{
			name:     "very long string",
			input:    strings.Repeat("a", 300),
			expected: strings.Repeat("a", 256),
		},
		{
			name:     "valid with spaces",
			input:    "test value with spaces",
			expected: "test value with spaces",
		},
		{
			name:     "valid with dots and hyphens",
			input:    "test.value-123",
			expected: "test.value-123",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeLabelValue(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeLabelValue(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty email",
			input:    "",
			expected: "unknown",
		},
		{
			name:     "normal email",
			input:    "user@example.com",
			expected: "user@example.com",
		},
		{
			name:     "email with special chars",
			input:    "user+tag@example.com",
			expected: "user+tag@example.com",
		},
		{
			name:     "very long email",
			input:    strings.Repeat("a", 300) + "@example.com",
			expected: (strings.Repeat("a", 300) + "@example.com")[:256],
		},
		{
			name:     "email with injection attempt",
			input:    "user<script>@example.com",
			expected: "user_script_@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeEmail(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeEmail(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeHostname(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty hostname",
			input:    "",
			expected: "unknown",
		},
		{
			name:     "normal hostname",
			input:    "server-01.example.com",
			expected: "server-01.example.com",
		},
		{
			name:     "uppercase hostname",
			input:    "SERVER-01.EXAMPLE.COM",
			expected: "server-01.example.com",
		},
		{
			name:     "hostname with spaces",
			input:    "my server",
			expected: "my server",
		},
		{
			name:     "hostname with special chars",
			input:    "server<script>",
			expected: "server_script_",
		},
		{
			name:     "very long hostname",
			input:    strings.Repeat("a", 300),
			expected: strings.Repeat("a", 256),
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeHostname(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeHostname(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
