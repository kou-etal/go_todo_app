// アプリ共通のユーティリティ、大規模の場合internal/pkg/clock/clock.goに置く、loggerとかmetricsとか
// 見せないというより切り替えのinterface
package clock

import (
	"time"
)

type Clocker interface {
	Now() time.Time
}

type RealClocker struct{}

func (r RealClocker) Now() time.Time {
	return time.Now()
}

type FixedClocker struct{}

func (fc FixedClocker) Now() time.Time {
	return time.Date(2025, 5, 10, 12, 34, 56, 0, time.UTC)
}
