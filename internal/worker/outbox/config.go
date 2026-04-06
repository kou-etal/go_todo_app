package outbox

import "time"

type Config struct {
	IdleSleep         time.Duration
	ChunkMaxRows      int
	ChunkMaxBytes     int64
	LeaseDuration     time.Duration
	HeartbeatInterval time.Duration
	MaxAttempt        uint32
	BackoffBase       time.Duration
}

func DefaultConfig() Config {
	return Config{
		IdleSleep:         10 * time.Second,
		ChunkMaxRows:      500,
		ChunkMaxBytes:     5 * 1024 * 1024,
		LeaseDuration:     60 * time.Second,
		HeartbeatInterval: 20 * time.Second,
		MaxAttempt:        5,
		BackoffBase:       1 * time.Minute,
	}
}
