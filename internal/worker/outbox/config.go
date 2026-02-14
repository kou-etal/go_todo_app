package outbox

import "time"

//ここはworker系に関する設定やからinternal/config(go-toto-app系の設定)とは分ける。
/*
defaultは定義しといてあとでenvから読み込む。現在はdefaultだけ。
cfg := DefaultConfig()
overrideFromEnv(&cfg)
読み取るときは&cfg
返すときは値かポインタどっちでもいい。appと違って共有多くない、サイズ大きくならない、動的ではない(immutable)から値で良い。
*/
//計算とか不等号を前提とするやつはtime.Timeじゃなくてtime.Durationで定義する。
//そうではなく単なるじかんを表すときはtime.Time、きかんを表すときはtime.Duration

//大きくなりそうな数値はint64
type Config struct {
	IdleSleep         time.Duration
	ChunkMaxRows      int
	ChunkMaxBytes     int64
	LeaseDuration     time.Duration //これはlease_untilで使う。now.Add(leaseDuration)
	HeartbeatInterval time.Duration //HeartbeatInterval<LeaseDurationにしないと切れるのに延長されない。
	MaxAttempt        uint32        //実際はuint32も使わないがunsigned intと比べやすいようにuint32
	BackoffBase       time.Duration
}

func DefaultConfig() Config {
	return Config{
		IdleSleep:         10 * time.Second, //time.Durationの定義方法
		ChunkMaxRows:      500,
		ChunkMaxBytes:     5 * 1024 * 1024, // 5MB
		LeaseDuration:     60 * time.Second,
		HeartbeatInterval: 20 * time.Second,
		MaxAttempt:        5,
		BackoffBase:       1 * time.Minute,
	}
}
