// アプリ共通のユーティリティ、大規模の場合internal/pkg/clock/clock.goに置く、loggerとかmetricsとか
// utils/を作るのは便利箱になってあまり良くない。
// 見せないというより切り替えのinterface
package clock

import (
	"time"
)

// Newは依存注入、設定、命名に分岐を委譲する目的。
// usecaseとかのnewは依存注入、loggerのnewは命名に分岐委譲(newslog()で引数なし)
// 今回はそれらなし

//appでclk定義、clkはclocker型やからusecaseはclk.Nowするだけやけど実際はapp次第で挙動が変わる->ポリモーフィズム
// Goは継承ではなくinterfaceのポリモーフィズム実現

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
