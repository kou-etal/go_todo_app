package outbox

import (
	"math"
	"math/rand/v2"
	"time"
)

//失敗した場合に使う。
//jitter採用。
// NextAttemptAt は指数バックオフ + Equal Jitter で次回リトライ時刻を算出する。
// attempt は今回の試行回数（1-indexed: 1回目の失敗 → attempt=1）。
//
// Equal Jitter: exp = base * 2^attempt, half = exp/2, jitter ∈ [0, half)
// 結果は [half, exp) の範囲に収まる。
//
// | attempt | 範囲          |
// |---------|---------------|
// | 1       | 30s 〜 60s    |
// | 2       | 1m 〜 2m      |
// | 3       | 2m 〜 4m      |
// | 4       | 4m 〜 8m      |
// | 5       | → DLQ         |

/*
time.Second
time.Minute
time.Hour
はtime.Duration型
type Duration int64
*/
func NextAttemptAt(now time.Time, attempt int, base time.Duration) time.Time {
	exp := base * time.Duration(math.Pow(2, float64(attempt)))
	//math.powはfloat64やからtime.Durationにキャストして計算する。math.Powは遅い。

	half := exp / 2
	jitter := time.Duration(rand.Int64N(int64(half))) //durationがint型やからfloatじゃなくてintに合わせるが楽。大きい数になる可能性あるからint64にキャスト。
	return now.Add(half + jitter)
}
