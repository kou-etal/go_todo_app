## 1.ドメイン・ユースケース
### エンティティ
#### Task
概要　ユーザーが扱う作業<br>
- 属性:
    - id
    - title
    - description
    - status (todo / doing / done)
    - dueDate
    - created_at
    - updated_at
#### User
概要　ユーザー情報<br>
- 属性:
    - id
    - name
    - email
    - password
    - is_admin
    - loggin_failed_count
    - is_banned
### ユースケース
- Actor:ユーザー
    - ユーザー情報を登録する
    - ログインする
    - タスクを作成し、更新し、一覧を見る
- Actor:admin
    - タスクを削除する
### ビジネスルール
- エンティティに対するルール
    - Task
        - title 必須、20文字以内
        - description 必須、1000文字まで
        - status 必須、デフォルトはtodo
        - dueDate 必須、1年まで
    - User
        - name 必須、10文字目まで
        - email 必須、@を含む
        - password 必須、10文字以上、大文字、小文字、記号をすべて含む
        - is_admin デフォルトは0
- ユースケースに対するルール
    - Task
        - task作成　作成数は10個まで
        - task一覧　期限が来ると一覧から消える
    - User
        - ログイン　5回連続で失敗すると一日ログイン拒否

## 2.API設計
### REST API
  認可ラベル:
- all: 未ログインでも利用可能
- login: ログイン済みユーザーのみ
- admin: admin権限を持つユーザーのみ

| 機能 | Method | Path | Request | Response | 認可 |
|-----|---------|------|---------|----------|-----|
|タスク一覧 | GET    |/tasks      |none	                                |Task[]| login|
|タスク作成 | POST   |/tasks      |title/description/due_date　　　　   |Task  | login|
|タスク更新 | PATCH  |/tasks/{id} |title?/description?/status?/due_date?|Task  | login|
|タスク削除 | DELETE |/tasks/{id} |none                                 |204  | admin|
|ユーザー登録| POST  |/users      |name/email/password                  | name/201| all  |
|ユーザーログイン|POST|/auth      |email/password                       |access_token/refresh_token  | all  |
### エラー設計

#### エラー伝搬の方針
各レイヤーが自パッケージのセンチネルエラーのみを返し、上位レイヤーが下位を import しない設計。

```
Domain 層:  dtask.ErrEmptyTitle, dtask.ErrConflict, duser.ErrConflict ...
               ↓ (errors.Is で吸収)
Usecase 層: create.ErrEmptyTitle, update.ErrConflict, register.ErrConflict ...
               ↓ (errors.Is で分岐)
Handler 層: HTTP ステータスコード + フィールド別メッセージ
```

- Handler は usecase パッケージのエラーのみを参照する。domain パッケージを import しない
- Usecase は domain エラーを `errors.Is()` で判定し、自パッケージのセンチネルに変換して返す
- すべてのエラーは `errors.New()` で定義し、`errors.Is()` で型安全に判定

#### HTTP エラーレスポンス形式
```json
{
  "message": "title is required"
}
```
- 400: バリデーションエラー（フィールド別メッセージ: "title is required", "invalid due_date" 等）
- 404: リソース未検出
- 409: 楽観ロック競合
- 500: 内部エラー（メッセージは "internal server error" 固定、詳細はログへ）

#### ログレベルの使い分け
- 4xx 系: `Debug` レベル（クライアント起因のため、アラート不要）
- 5xx 系: `Error` レベル（サーバー起因のため、アラート対象）

## 3.アーキテクチャ
クリーンアーキテクチャを採用
- cmd: エントリポイント（サーバー起動）
- handler(presentation): HTTP層（リクエスト/レスポンスの整形）
- usecase: ユースケース実行ロジック（アプリケーションサービス）
- domain: エンティティ、Value Object、ビジネスロジックの実装
- infra: 外部I/O（DB・外部API等）の具体実装
- app: DI コンテナ（全レイヤーの組み立て）

依存関係のルール
- domain は他レイヤーに依存しない
- handler は usecase に依存する（domain を import しない）
- usecase は domain に依存するが、infra/handler には依存しない
- infra は domain に依存する（domain の Repository interface を実装）
- cmd/app は全レイヤーを束ねるが、ビジネスロジックは持たない

### Interface の配置
- **Repository interface** → domain 層に定義（`domain/task/repository.go`, `domain/user/repository.go`）
  - infra 層が domain の interface を実装する（依存性逆転）
- **Logger interface** → `internal/logger/` に定義（横断的関心事のため独立パッケージ）
- **Usecase** → interface を置かず、Handler が具体型 `*Usecase` に直接依存
  - 理由: Handler のテストは薄い（mock usecase ではなく、実 usecase + mock repo で構成）
- **Worker 依存** → worker パッケージ内に interface 定義（`outbox.OutboxRepo`, `outbox.ObjectUploader`）
  - 理由: Worker は usecase 経由せず infra に直接依存するため、使う側で定義

### トランザクション設計
`Runner[D any]` ジェネリクスによるトランザクション境界の抽象化。

```go
// usecase/tx/runner.go
type Runner[D any] interface {
    WithinTx(ctx context.Context, fn func(ctx context.Context, deps D) error) error
}
```

- `D` はトランザクション内で使う依存群（Deps interface）
  - `TaskEventDeps`: TaskRepo + TaskEventRepo（タスク CRUD 用）
  - `RegisterDeps`: UserRepo + EmailVerifyRepo（ユーザー登録用）
- `infra/db/tx/runner.go` に `SQLxRunner[D]` として具体実装
  - `BeginTxx` → `fn(ctx, deps)` → `Commit` / `Rollback`（panic recovery 付き）
- app 層で `DepsFactory[D]` 関数を渡してトランザクション内の deps を組み立て

利点:
- usecase は `*sqlx.Tx` を知らない（DB 非依存）
- トランザクション内のリポジトリ構成を app 層で自由に差し替え可能
- テスト時は mock runner を差し込むだけ

### 楽観ロック
Task / User エンティティに `version` カラム（BIGINT UNSIGNED）を持たせる。

```
1. FindByID で現在の version を取得
2. usecase 側でも cmd.Version と比較（アプリレベル CAS）
3. UPDATE ... WHERE id = ? AND version = ? （SQL レベル CAS）
4. affected_rows == 0 → ErrConflict
```

- 更新・削除の両方で適用
- Handler は 409 Conflict を返す

### Middleware チェーン

```
HTTP Request
  → AccessLog（レスポンス時間・ステータス・バイト数の記録）
    → RequestID（X-Request-Id 生成/伝搬）
      → Router → Handler → Usecase → Domain
```

- `responseRecorder`: `http.ResponseWriter` をラップし statusCode / bytes を観測
- RequestID: 既存ヘッダーがあれば尊重、なければ UUID 生成。`context.WithValue` で伝搬
- AccessLog: 500 系は `Error`、それ以外は `Info` レベルでログ出力
- clientIP: `X-Forwarded-For` → `X-Real-IP` → `RemoteAddr` の優先順で取得

## 4.イベントパイプライン

### Transactional Outbox パターン
タスク操作（create/update/delete）のイベントを、ビジネストランザクション内で `task_events` テーブルに書き込む。
外部配信はトランザクション外で非同期に行う。

```
[API Server]
  Usecase: タスク操作 + イベント挿入を同一 Tx で実行
      ↓ (task_events テーブル)

[Outbox Worker] (cmd/event-worker)
  ループ:
    1. Claim: FOR UPDATE SKIP LOCKED で未配信イベントを取得
    2. SetLease: リース取得（owner + 有効期限）
    3. Heartbeat: 別 goroutine でリースを延長（CAS: lease_until 一致が条件）
    4. emitToS3: JSONL + manifest.json を S3 にアップロード
       - manifest が既に存在する場合は冪等スキップ
    5. MarkEmitted: emitted_at を記録（配信完了）
    6. 失敗時: attempt_count に応じてリトライ or DLQ 移動
      ↓ (S3: raw/task-events/year=.../month=.../day=.../hour=.../...)

[Compaction Worker] (cmd/compaction)
  日次バッチ:
    1. S3 から JSONL ファイルを読み込み
    2. EventID ベースで重複排除（dedupe）
    3. 日付ごとに Apache Parquet に変換
    4. compaction manifest + .done マーカーで冪等性を担保
      ↓ (S3: compacted/task-events/day=.../...)
```

### Outbox Worker 詳細

#### リース機構
- 複数 Worker インスタンスが安全に並行動作するための排他制御
- `lease_owner`: Worker の UUID。自分が取得したイベントのみ操作可能
- `lease_until`: リース有効期限。超過すると他 Worker が奪取可能
- Heartbeat: 20秒間隔でリースを延長。CAS（`WHERE lease_until = ?`）で整合性を保証
- CAS mismatch → リース喪失と判断し、処理を中断

#### リトライ・DLQ
- 指数バックオフ: `base * 2^attempt` の半分 + jitter（thundering herd 防止）
- デフォルト: base=1分、最大5回
- 5回失敗 → `task_events_dlq` テーブルに移動（last_error を記録）

#### 冪等性
- S3 アップロード前に manifest の存在チェック
- manifest が既に存在 → アップロードスキップ → MarkEmitted のみ実行
- batch_id: イベントID群 + claimed_at + schema_version の SHA-256 ハッシュ

#### 設定値
| パラメータ | デフォルト | 説明 |
|-----------|----------|------|
| IdleSleep | 10秒 | Claim 0件時のスリープ |
| ChunkMaxRows | 500 | 1バッチの最大イベント数 |
| ChunkMaxBytes | 5MB | 1バッチの最大ペイロードサイズ |
| LeaseDuration | 60秒 | リース有効期間 |
| HeartbeatInterval | 20秒 | リース延長間隔 |
| MaxAttempt | 5 | 最大試行回数 |
| BackoffBase | 1分 | バックオフ基準時間 |

### Compaction Worker 詳細

#### 処理フロー
1. 対象日のイベントを S3 (raw/) から読み込み
2. BackfillWindow（デフォルト2日分）のバッファ期間を含めて収集
3. EventID ベースで重複排除
4. occurred_at の日付でグルーピング
5. 日ごとに Apache Parquet ファイルを生成・アップロード
6. compaction manifest を書き込み
7. .done マーカーを書き込み（冪等性: 次回実行時にスキップ）

#### S3 キー設計
```
raw/task-events/year=2026/month=02/day=01/hour=12/{batch_id}.jsonl
raw/task-events/year=2026/month=02/day=01/hour=12/{batch_id}.manifest.json
compacted/task-events/day=2026-02-01/claim_2026-02-01.parquet
_state/compaction/2026-02-01.manifest.json
_state/compaction/2026-02-01.done
```

## 5.データベース設計

### テーブル一覧

#### users
ユーザー情報を管理する。

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | CHAR(36) | PK | UUID |
| email | VARCHAR(254) | UNIQUE | メールアドレス（小文字正規化） |
| password_hash | VARCHAR(80) | NOT NULL | bcrypt ハッシュ |
| user_name | VARCHAR(20) | UNIQUE | ユーザー名 |
| email_verified_at | DATETIME(6) | NULL | メール認証日時 |
| version | BIGINT UNSIGNED | NOT NULL | 楽観ロック |
| created_at / updated_at | DATETIME(6) | NOT NULL | タイムスタンプ |

#### task
ユーザーのタスクを管理する。

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | CHAR(36) | PK | UUID |
| user_id | CHAR(36) | FK → users.id | 所有者 |
| title | VARCHAR(20) | NOT NULL | タイトル |
| description | VARCHAR(1000) | NOT NULL | 説明 |
| status | VARCHAR(20) | NOT NULL | todo / doing / done |
| due_date | DATETIME(6) | NOT NULL | 期限 |
| version | BIGINT UNSIGNED | NOT NULL | 楽観ロック |
| due_is_null | TINYINT(1) | STORED | インデックス用生成カラム |

インデックス:
- `idx_task_user_created_sort (user_id, created_at, id)` — 作成日ソート用
- `idx_task_user_due_sort (user_id, due_is_null, due_date, id)` — 期限ソート用

#### email_verification_tokens
メール認証トークンを管理する。

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | CHAR(36) | PK | UUID |
| user_id | CHAR(36) | FK → users.id | 対象ユーザー |
| token_hash | CHAR(64) | UNIQUE | SHA-256(plain_token) の hex |
| expires_at | DATETIME(6) | NOT NULL | 有効期限 |
| used_at | DATETIME(6) | NULL | 使用日時 |

#### task_events（Outbox テーブル）
タスク操作イベントを蓄積し、Outbox Worker が外部配信する。

| カラム | 型 | 制約 | 説明 |
|-------|-----|------|------|
| id | CHAR(36) | PK | イベントID |
| user_id / task_id / request_id | CHAR(36) | NOT NULL | トレーシング用 |
| event_type | VARCHAR(20) | NOT NULL | created / updated / deleted |
| occurred_at | DATETIME(6) | NOT NULL | イベント発生日時 |
| emitted_at | DATETIME(6) | NULL | 配信完了日時（未配信は NULL） |
| schema_version | INT UNSIGNED | NOT NULL | ペイロードスキーマバージョン |
| payload | JSON | NOT NULL | イベントペイロード |
| next_attempt_at | DATETIME(6) | NOT NULL | 次回リトライ可能時刻 |
| attempt_count | INT UNSIGNED | NOT NULL | 試行回数 |
| lease_owner | VARCHAR(255) | NULL | リース保持 Worker ID |
| lease_until | DATETIME(6) | NULL | リース有効期限 |
| claimed_at | DATETIME(6) | NULL | Claim 時刻 |

インデックス:
- `idx_task_events_claimable (emitted_at, next_attempt_at, lease_until)` — Claim クエリ用

#### task_events_dlq
配信失敗イベントの退避先（Dead Letter Queue）。

| カラム | 型 | 説明 |
|-------|-----|------|
| id | CHAR(36) | 元イベントID |
| last_error | TEXT | 最後のエラーメッセージ |
| dead_at | DATETIME(6) | DLQ 投入日時 |
| （その他） | | task_events と同一構造 |

## 6.可観測性

### 構造化ログ
- `internal/logger/` に `Logger` interface を定義
- 具体実装は `log/slog`（`logger/slog.go`）
- メソッド: `Debug`, `Info`, `Warn`, `Error`
- `logger.Attr` で Key-Value を渡す。slog/zap/zerolog を差し替え可能
- Request ID は `context.WithValue` で伝搬し、Logger の具体実装で付与

### Prometheus メトリクス

#### Outbox Worker（16 メトリクス）
| 種別 | メトリクス名 | 説明 |
|------|------------|------|
| Histogram | outbox_process_duration_seconds | 1サイクルの処理時間 |
| Histogram | outbox_claim_duration_seconds | Claim (SELECT FOR UPDATE) の所要時間 |
| Histogram | outbox_claim_batch_size | Claim で取得したレコード数 |
| Histogram | outbox_upload_data_duration_seconds | JSONL の S3 PUT 所要時間 |
| Histogram | outbox_upload_manifest_duration_seconds | manifest の S3 PUT 所要時間 |
| Histogram | outbox_heartbeat_extend_duration_seconds | ExtendLease の所要時間 |
| Histogram | outbox_event_lag_to_claim_seconds | occurred_at → claim の遅延 |
| Histogram | outbox_event_lag_to_emit_seconds | occurred_at → emit の遅延 |
| Counter | outbox_events_emitted_total | emit 成功イベント数 |
| Counter | outbox_events_retried_total | リトライに回ったイベント数 |
| Counter | outbox_events_dlq_total | DLQ 移動イベント数 |
| Counter | outbox_repo_failures_total | DB 操作失敗数（ラベル: op） |
| Counter | outbox_heartbeat_lost_total | リース喪失回数 |
| Counter | outbox_idle_cycles_total | Claim 0件のアイドルサイクル |
| Gauge | outbox_inflight_leased | 現在リース中の件数 |
| Gauge | outbox_queue_depth | 未 emit のキュー深度 |

#### Compaction Worker（14 メトリクス）
| 種別 | メトリクス名 | 説明 |
|------|------------|------|
| Gauge | compaction_last_success_timestamp | 最後の成功時刻 |
| Gauge | compaction_last_failure_timestamp | 最後の失敗時刻 |
| Gauge | compaction_last_run_status | 直近の実行結果 (1=成功, 0=失敗) |
| Gauge | compaction_last_run_duration_seconds | 直近の実行所要時間 |
| Gauge | compaction_events_read | 読み込みイベント数 |
| Gauge | compaction_events_deduped | dedupe 後のイベント数 |
| Gauge | compaction_parquet_files | 生成 Parquet ファイル数 |
| Gauge | compaction_s3_list_duration_seconds | S3 List の合計所要時間 |
| Gauge | compaction_s3_read_duration_seconds | S3 Get (JSONL 読み込み) の所要時間 |
| Gauge | compaction_parquet_write_duration_seconds | Parquet 変換の所要時間 |
| Gauge | compaction_s3_upload_duration_seconds | S3 Upload (Parquet) の所要時間 |

### Grafana
- `docker/grafana/provisioning/` にデータソース・ダッシュボードのプロビジョニング設定を同梱
- `docker/grafana/dashboards/` にダッシュボード JSON を配置
- Prometheus → Grafana の自動接続

## 7.テスト方針

### 方針
- 外部ライブラリ不使用: testify 等を使わず、標準 `testing` パッケージのみ
- 手書き mock / stub: コード生成ツール不使用。テスト対象と同パッケージに定義
- テーブル駆動テスト + `t.Parallel()` で並列実行
- `t.Run()` でサブテスト化し、テスト名を明示

### テスト構成

| レイヤー | テスト対象 | テスト手法 |
|---------|----------|----------|
| Domain | 値オブジェクト生成・バリデーション | テーブル駆動テスト。依存ゼロ |
| Domain | エンティティのビジネスルール | ヘルパーで reconstruct → メソッド呼び出し |
| Usecase | ユースケースの正常系・異常系 | mock runner + mock repo を注入 |
| Handler | HTTP ステータスコード・レスポンスボディ | httptest + 実 usecase (mock repo 注入) |
| Middleware | RequestID 生成・伝搬 | httptest + inner handler で context 検証 |
| Middleware | responseRecorder | WriteHeader / Write の状態検証 |
| Responder | JSON レスポンス整形 | httptest で Content-Type・ボディ検証 |
| Worker | Outbox / Compaction の処理フロー | mock repo + mock uploader |

### Handler テストの設計判断
Handler は usecase の具体型 `*Usecase` に直接依存している（interface 不使用）。
テスト時は mock usecase ではなく、実 usecase に mock runner / mock repo を注入して HTTP レベルで検証する。

理由:
- Handler のロジックは薄い（JSON パース → usecase 呼び出し → エラー分岐 → レスポンス返却）
- usecase ごとに interface を定義するコストに見合わない
- 実 usecase を通すことで、バリデーション込みの結合テストになる
