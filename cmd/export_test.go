package cmd

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		hasError bool
	}{
		{
			name:     "YYYY-MM-DD format",
			input:    "2024-01-15",
			expected: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			hasError: false,
		},
		{
			name:     "YYYY-MM-DD HH:MM:SS format",
			input:    "2024-01-15 14:30:45",
			expected: time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC),
			hasError: false,
		},
		{
			name:     "ISO format with T",
			input:    "2024-01-15T14:30:45",
			expected: time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC),
			hasError: false,
		},
		{
			name:     "ISO format with Z",
			input:    "2024-01-15T14:30:45Z",
			expected: time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC),
			hasError: false,
		},
		{
			name:     "Invalid format",
			input:    "invalid-date",
			expected: time.Time{},
			hasError: true,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: time.Time{},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDate(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input '%s', but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input '%s': %v", tt.input, err)
				}

				if !result.Equal(tt.expected) {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestGetStageEmoji(t *testing.T) {
	tests := []struct {
		stage    string
		expected string
	}{
		{"initializing", "🚀"},
		{"channel_fetch", "📋"},
		{"message_fetch", "💬"},
		{"thread_fetch", "🧵"},
		{"user_fetch", "👥"},
		{"data_processing", "⚙️"},
		{"file_generation", "📁"},
		{"complete", "✅"},
		{"unknown", "🔄"},
	}

	for _, tt := range tests {
		t.Run(tt.stage, func(t *testing.T) {
			result := getStageEmoji(tt.stage)
			if result != tt.expected {
				t.Errorf("Expected emoji '%s' for stage '%s', got '%s'", tt.expected, tt.stage, result)
			}
		})
	}
}

func TestGetStageDescription(t *testing.T) {
	tests := []struct {
		stage    string
		expected string
	}{
		{"initializing", "Initializing export"},
		{"channel_fetch", "Fetching channel information"},
		{"message_fetch", "Fetching messages"},
		{"thread_fetch", "Fetching thread replies"},
		{"user_fetch", "Fetching user information"},
		{"data_processing", "Processing data"},
		{"file_generation", "Generating output file"},
		{"complete", "Export complete"},
		{"unknown", "Processing"},
	}

	for _, tt := range tests {
		t.Run(tt.stage, func(t *testing.T) {
			result := getStageDescription(tt.stage)
			if result != tt.expected {
				t.Errorf("Expected description '%s' for stage '%s', got '%s'", tt.expected, tt.stage, result)
			}
		})
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "Bytes",
			bytes:    512,
			expected: "512 B",
		},
		{
			name:     "Kilobytes",
			bytes:    1536, // 1.5 KB
			expected: "1.5 KB",
		},
		{
			name:     "Megabytes",
			bytes:    2097152, // 2 MB
			expected: "2.0 MB",
		},
		{
			name:     "Gigabytes",
			bytes:    3221225472, // 3 GB
			expected: "3.0 GB",
		},
		{
			name:     "Zero bytes",
			bytes:    0,
			expected: "0 B",
		},
		{
			name:     "One byte",
			bytes:    1,
			expected: "1 B",
		},
		{
			name:     "Exactly 1 KB",
			bytes:    1024,
			expected: "1.0 KB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFileSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("Expected '%s' for %d bytes, got '%s'", tt.expected, tt.bytes, result)
			}
		})
	}
}
