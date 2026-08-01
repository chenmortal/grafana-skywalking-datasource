package plugin

import (
	"testing"
	"time"
)

func TestFormatTimeToString(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		step     Step
		expected string
	}{
		{
			name:     "StepDay - basic date",
			input:    time.Date(2026, 6, 2, 2, 46, 30, 0, time.UTC),
			step:     StepDay,
			expected: "2026-06-02",
		},
		{
			name:     "StepHour - date and hour",
			input:    time.Date(2026, 6, 2, 2, 46, 30, 0, time.UTC),
			step:     StepHour,
			expected: "2026-06-02 02",
		},
		{
			name:     "StepMinute - date, hour and minute",
			input:    time.Date(2026, 6, 2, 2, 46, 30, 0, time.UTC),
			step:     StepMinute,
			expected: "2026-06-02 0246",
		},
		{
			name:     "default - full datetime",
			input:    time.Date(2026, 6, 2, 2, 46, 30, 0, time.UTC),
			step:     Step("UNKNOWN"),
			expected: "2026-06-02 024630",
		},
		{
			name:     "StepDay - midnight",
			input:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			step:     StepDay,
			expected: "2026-01-01",
		},
		{
			name:     "StepHour - end of day",
			input:    time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
			step:     StepHour,
			expected: "2026-12-31 23",
		},
		{
			name:     "StepMinute - morning",
			input:    time.Date(2026, 6, 2, 10, 15, 0, 0, time.UTC),
			step:     StepMinute,
			expected: "2026-06-02 1015",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTimeToString(tt.input, tt.step)
			if result != tt.expected {
				t.Errorf("formatTimeToString(%v, %v) = %v, want %v", tt.input, tt.step, result, tt.expected)
			}
		})
	}
}

func BenchmarkFormatTimeToString(b *testing.B) {
	t := time.Date(2026, 6, 2, 2, 46, 30, 0, time.UTC)
	steps := []Step{StepDay, StepHour, StepMinute}

	
	for i := 0; b.Loop(); i++ {
		formatTimeToString(t, steps[i%len(steps)])
	}
}
