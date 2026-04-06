package compaction

type Config struct {
	BackfillWindow  int
	RawPrefix       string
	CompactedPrefix string
	StatePrefix     string
	S3Bucket        string
	S3Endpoint      string
}

func DefaultConfig() Config {
	return Config{
		BackfillWindow:  2,
		RawPrefix:       "raw/task-events",
		CompactedPrefix: "compacted/task-events",
		StatePrefix:     "_state/compaction",
	}
}
