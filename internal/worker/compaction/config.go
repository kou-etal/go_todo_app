package compaction

//compaction周りのデフォルト設定。デプロイ環境ではDI層からenv使って代入する。
//これは主にローカル設定。
//デファクト

type Config struct {
	BackfillWindow  int    //  raw を読む範囲。days
	RawPrefix       string // raw ファイルのプレフィックス
	CompactedPrefix string // 出力先プレフィックス
	StatePrefix     string // 完了マーカーのプレフィックス
	S3Bucket        string
	S3Endpoint      string
}

func DefaultConfig() Config {
	return Config{
		BackfillWindow:  2,
		RawPrefix:       "raw/task-events",
		CompactedPrefix: "compacted/task-events",
		StatePrefix:     "_state/compaction",
		//デフォルトでS3Bucket、S3Endpoint代入してないからDIでバリデーション
	}
}
