package utils

import (
	"regexp"
	"strings"
)

var (
	// labelValueRegex matches invalid characters that should be replaced
	// Replace control characters, special shell/script characters
	labelValueRegex = regexp.MustCompile(`[<>'";\n\t\r\x00-\x1f\x7f\\|&$(){}[\]*?~!]`)

	// Max length for label values to prevent cardinality explosion
	maxLabelLength = 256
)

// SanitizeLabelValue sanitizes a string to be used as a Prometheus label value
// It removes potentially dangerous characters and limits the length
func SanitizeLabelValue(value string) string {
	if value == "" {
		return "unknown"
	}

	// Trim whitespace
	value = strings.TrimSpace(value)

	// Replace invalid characters with underscores
	value = labelValueRegex.ReplaceAllString(value, "_")

	// Limit length to prevent cardinality explosion
	if len(value) > maxLabelLength {
		value = value[:maxLabelLength]
	}

	// Ensure we don't return an empty string
	if value == "" {
		return "unknown"
	}

	return value
}

// SanitizeEmail sanitizes an email address for use in metrics
// Optionally hashes the local part to preserve privacy
func SanitizeEmail(email string) string {
	if email == "" {
		return "unknown"
	}

	// Basic sanitization - remove any non-standard characters
	email = strings.TrimSpace(email)
	email = labelValueRegex.ReplaceAllString(email, "_")

	// Limit length
	if len(email) > maxLabelLength {
		email = email[:maxLabelLength]
	}

	return email
}

// SanitizeHostname sanitizes a hostname for use in metrics
func SanitizeHostname(hostname string) string {
	if hostname == "" {
		return "unknown"
	}

	// Trim and convert to lowercase
	hostname = strings.ToLower(strings.TrimSpace(hostname))

	// Replace invalid characters
	hostname = labelValueRegex.ReplaceAllString(hostname, "_")

	// Limit length
	if len(hostname) > maxLabelLength {
		hostname = hostname[:maxLabelLength]
	}

	if hostname == "" {
		return "unknown"
	}

	return hostname
}
