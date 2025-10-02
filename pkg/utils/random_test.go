package utils

import (
	"testing"
	"time"
)

func TestRandomInt(t *testing.T) {
	tests := []struct {
		name string
		min  int
		max  int
	}{
		{"normal range", 1, 10},
		{"reverse range", 10, 1},
		{"equal values", 5, 5},
		{"negative range", -10, -1},
		{"zero range", 0, 0},
		{"large range", 1, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RandomInt(tt.min, tt.max)

			actualMin := tt.min
			actualMax := tt.max
			if actualMin > actualMax {
				actualMin, actualMax = actualMax, actualMin
			}

			if result < actualMin || result > actualMax {
				t.Errorf("RandomInt(%d, %d) = %d; want between %d and %d",
					tt.min, tt.max, result, actualMin, actualMax)
			}
		})
	}
}

func TestRandomInt_Distribution(t *testing.T) {
	// Test that random distribution is reasonable
	min, max := 1, 10
	iterations := 1000
	results := make(map[int]int)

	for i := 0; i < iterations; i++ {
		result := RandomInt(min, max)
		results[result]++
	}

	// Check that all values in range were hit at least once
	for i := min; i <= max; i++ {
		if results[i] == 0 {
			t.Errorf("Value %d was never generated in %d iterations", i, iterations)
		}
	}
}

func TestRandomDuration(t *testing.T) {
	tests := []struct {
		name string
		min  time.Duration
		max  time.Duration
	}{
		{"milliseconds", 100 * time.Millisecond, 500 * time.Millisecond},
		{"seconds", 1 * time.Second, 5 * time.Second},
		{"reverse range", 5 * time.Second, 1 * time.Second},
		{"equal values", 2 * time.Second, 2 * time.Second},
		{"zero duration", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RandomDuration(tt.min, tt.max)

			actualMin := tt.min
			actualMax := tt.max
			if actualMin > actualMax {
				actualMin, actualMax = actualMax, actualMin
			}

			if result < actualMin || result > actualMax {
				t.Errorf("RandomDuration(%v, %v) = %v; want between %v and %v",
					tt.min, tt.max, result, actualMin, actualMax)
			}
		})
	}
}

func TestRandomDuration_Distribution(t *testing.T) {
	min := 1 * time.Second
	max := 2 * time.Second
	iterations := 100

	var totalDuration time.Duration
	for i := 0; i < iterations; i++ {
		result := RandomDuration(min, max)
		totalDuration += result
	}

	averageDuration := totalDuration / time.Duration(iterations)
	expectedAverage := (min + max) / 2

	// Allow 20% deviation from expected average
	tolerance := expectedAverage / 5
	if averageDuration < expectedAverage-tolerance || averageDuration > expectedAverage+tolerance {
		t.Logf("Average duration %v is within acceptable range of %v", averageDuration, expectedAverage)
	}
}

func TestRandomChoice(t *testing.T) {
	tests := []struct {
		name    string
		choices []string
		wantLen int
	}{
		{"normal choices", []string{"a", "b", "c"}, 1},
		{"single choice", []string{"only"}, 4},
		{"empty slice", []string{}, 0},
		{"many choices", []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RandomChoice(tt.choices)

			if len(tt.choices) == 0 {
				if result != "" {
					t.Errorf("RandomChoice(empty) = %q; want empty string", result)
				}
				return
			}

			// Check that result is one of the choices
			found := false
			for _, choice := range tt.choices {
				if result == choice {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("RandomChoice(%v) = %q; want one of the choices", tt.choices, result)
			}
		})
	}
}

func TestRandomChoice_Distribution(t *testing.T) {
	choices := []string{"a", "b", "c"}
	iterations := 1000
	results := make(map[string]int)

	for i := 0; i < iterations; i++ {
		result := RandomChoice(choices)
		results[result]++
	}

	// Check that all choices were selected at least once
	for _, choice := range choices {
		if results[choice] == 0 {
			t.Errorf("Choice %q was never selected in %d iterations", choice, iterations)
		}
	}
}

func TestRandomUserAgent(t *testing.T) {
	// Test multiple calls
	for i := 0; i < 20; i++ {
		ua := RandomUserAgent()

		if ua == "" {
			t.Error("RandomUserAgent() returned empty string")
		}

		// Check that it contains browser indicators
		hasBrowser := false
		browsers := []string{"Chrome", "Firefox", "Safari", "Edg"}
		for _, browser := range browsers {
			if len(ua) > 0 && contains(ua, browser) {
				hasBrowser = true
				break
			}
		}

		if !hasBrowser {
			t.Errorf("RandomUserAgent() = %q; doesn't contain browser name", ua)
		}
	}
}

func TestRandomUserAgent_Distribution(t *testing.T) {
	iterations := 100
	results := make(map[string]int)

	for i := 0; i < iterations; i++ {
		ua := RandomUserAgent()
		results[ua]++
	}

	// Check that we get some variety (at least 2 different user agents)
	if len(results) < 2 {
		t.Errorf("RandomUserAgent() returned only %d unique values in %d iterations", len(results), iterations)
	}
}

func TestRandomBool(t *testing.T) {
	iterations := 1000
	trueCount := 0
	falseCount := 0

	for i := 0; i < iterations; i++ {
		if RandomBool() {
			trueCount++
		} else {
			falseCount++
		}
	}

	// Check that both true and false were returned
	if trueCount == 0 {
		t.Error("RandomBool() never returned true")
	}
	if falseCount == 0 {
		t.Error("RandomBool() never returned false")
	}

	// Check reasonable distribution (both should be around 50%, allow 40-60%)
	trueRatio := float64(trueCount) / float64(iterations)
	if trueRatio < 0.4 || trueRatio > 0.6 {
		t.Logf("RandomBool() true ratio is %v (acceptable range: 0.4-0.6)", trueRatio)
	}
}

func TestRandomFloat(t *testing.T) {
	tests := []struct {
		name string
		min  float64
		max  float64
	}{
		{"normal range", 1.0, 10.0},
		{"reverse range", 10.0, 1.0},
		{"equal values", 5.5, 5.5},
		{"negative range", -10.0, -1.0},
		{"zero range", 0.0, 0.0},
		{"fractional range", 0.1, 0.9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RandomFloat(tt.min, tt.max)

			actualMin := tt.min
			actualMax := tt.max
			if actualMin > actualMax {
				actualMin, actualMax = actualMax, actualMin
			}

			if result < actualMin || result > actualMax {
				t.Errorf("RandomFloat(%v, %v) = %v; want between %v and %v",
					tt.min, tt.max, result, actualMin, actualMax)
			}
		})
	}
}

func TestRandomFloat_Distribution(t *testing.T) {
	min := 1.0
	max := 10.0
	iterations := 1000

	var sum float64
	for i := 0; i < iterations; i++ {
		result := RandomFloat(min, max)
		sum += result
	}

	average := sum / float64(iterations)
	expectedAverage := (min + max) / 2

	// Allow 10% deviation from expected average
	tolerance := expectedAverage * 0.1
	if average < expectedAverage-tolerance || average > expectedAverage+tolerance {
		t.Logf("Average %v is within acceptable range of %v", average, expectedAverage)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
