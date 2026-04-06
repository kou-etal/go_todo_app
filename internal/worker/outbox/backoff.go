package outbox

import (
	"math"
	"math/rand/v2"
	"time"
)

func NextAttemptAt(now time.Time, attempt int, base time.Duration) time.Time {
	exp := base * time.Duration(math.Pow(2, float64(attempt)))
	//math.powはfloat64やからtime.Durationにキャストして計算する。math.Powは遅い。

	half := exp / 2
	jitter := time.Duration(rand.Int64N(int64(half)))
	return now.Add(half + jitter)
}
