package outbox

import (
	"testing"
	"time"
)

func TestNextAttemptAt_withinExpectedRange(t *testing.T) {
	t.Parallel()

	base := 1 * time.Minute
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name    string
		attempt int
		wantMin time.Duration // half = base * 2^attempt / 2
		wantMax time.Duration // exp  = base * 2^attempt
	}{
		{"attempt1", 1, 1 * time.Minute, 2 * time.Minute},
		{"attempt2", 2, 2 * time.Minute, 4 * time.Minute},
		{"attempt3", 3, 4 * time.Minute, 8 * time.Minute},
		{"attempt4", 4, 8 * time.Minute, 16 * time.Minute},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// 乱数の偏りを吸収するために複数回実行
			for range 100 {
				got := NextAttemptAt(now, tc.attempt, base)
				diff := got.Sub(now)

				if diff < tc.wantMin {
					t.Fatalf("attempt=%d: got %v, want >= %v", tc.attempt, diff, tc.wantMin)
				}
				if diff >= tc.wantMax {
					t.Fatalf("attempt=%d: got %v, want < %v", tc.attempt, diff, tc.wantMax)
				}
			}
		})
	}
}

func TestNextAttemptAt_alwaysAfterNow(t *testing.T) {
	t.Parallel()

	now := time.Now()
	base := 1 * time.Minute

	for attempt := 1; attempt <= 4; attempt++ {
		got := NextAttemptAt(now, attempt, base)
		if !got.After(now) {
			t.Fatalf("attempt=%d: result %v is not after now %v", attempt, got, now)
		}
	}
}

func TestNextAttemptAt_customBase(t *testing.T) {
	t.Parallel()

	base := 30 * time.Second
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// attempt=1, base=30s → exp=60s, half=30s → [30s, 60s)
	for range 100 {
		got := NextAttemptAt(now, 1, base)
		diff := got.Sub(now)

		if diff < 30*time.Second {
			t.Fatalf("got %v, want >= 30s", diff)
		}
		if diff >= 60*time.Second {
			t.Fatalf("got %v, want < 60s", diff)
		}
	}
}
