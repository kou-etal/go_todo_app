// utils/を作るのは便利箱になってあまり良くない。

package clock

import (
	"time"
)

type Clocker interface {
	Now() time.Time
}

type RealClocker struct{}

func (rc RealClocker) Now() time.Time {
	return time.Now()
}

type FixedClocker struct{}

func (fc FixedClocker) Now() time.Time {
	return time.Date(2025, 5, 10, 12, 34, 56, 0, time.UTC)
}

func NormalizeTime(t time.Time) time.Time {
	return t.UTC().Truncate(time.Second)
}
