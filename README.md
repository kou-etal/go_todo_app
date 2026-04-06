# Go Todo App

Goで構築した Todo REST API。
クリーンアーキテクチャをベースに、イベント駆動データパイプラインも実装しています。

## Architecture

```
cmd/
  todo-api/          ... HTTP API サーバー
  event-worker/      ... Outbox Worker (イベント配信)
  compaction/        ... Compaction Worker (Parquet 変換)

internal/
  domain/            ... エンティティ・ValueObject・リポジトリインターフェース
    task/            ... Task 集約 (Title, Description, Status, DueDate)
    user/            ... User 集約 (Email, Password, UserName)
    user/verification/ ... メール認証トークン
    event/           ... TaskEvent (created / updated / deleted)

  usecase/           ... アプリケーションサービス
    task/create/     ... タスク作成
    task/update/     ... タスク更新 
    task/delete/     ... タスク削除 
    task/list/       ... タスク一覧 
    user/register/   ... ユーザー登録 + メール認証トークン発行
    tx/              ... トランザクション境界 

  presentation/http/ ... HTTP ハンドラー・ミドルウェア・ルーター
  infra/             ... MySQL リポジトリ・S3 アップローダー・bcrypt 等
  worker/outbox/     ... Outbox Worker (Claim → Lease → S3 配信 → DLQ)
  worker/compaction/ ... Compaction Worker (JSONL → 重複排除 → Parquet)
```

### レイヤー依存の方向

```
Handler → Usecase → Domain ← Infra
                             
```

- **Domain 層**: 外部依存ゼロ。値オブジェクトでバリデーション、エンティティでビジネスルールを表現
- **Usecase 層**: ドメインエラーを吸収し、自パッケージのセンチネルエラーのみを返却。Handler がドメインを import しない設計
- **Handler 層**: Usecase エラーを HTTP ステータスコード + フィールド別メッセージに変換
- **Infra 層**: MySQL (sqlx) / S3 (aws-sdk-go-v2) / bcrypt / SHA-256 の具体実装

## Features

### REST API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | ヘルスチェック |
| `POST` | `/tasks` | タスク作成 |
| `GET` | `/tasks` | タスク一覧 |
| `PATCH` | `/tasks/{id}` | タスク更新 |
| `DELETE` | `/tasks/{id}` | タスク削除 |
| `POST` | `/users` | ユーザー登録 |

### Event Pipeline (Transactional Outbox)

```
[API] ──(Tx)--> task_events テーブル
                     │
              [Outbox Worker]
                  Claim (FOR UPDATE SKIP LOCKED)
                  → Lease 取得 (CAS + Heartbeat)
                  → S3 配信 (JSONL + manifest)
                  → MarkEmitted
                  → 失敗時: リトライ (指数バックオフ) or DLQ
                     │
             [Compaction Worker]
                  S3 JSONL 読み込み
                  → 重複排除 (EventID ベース)
                  → Apache Parquet 変換
                  → 冪等マーカー (.done)
```

### Observability

- **Prometheus メトリクス**: Outbox / Compaction Worker の処理時間、キュー深度、リトライ/DLQ カウント
- **Grafana ダッシュボード**: `docker/grafana/` に設定
- **構造化ログ**: `log/slog` ベースの Logger インターフェース。Request ID をコンテキスト伝搬
- **Request ID ミドルウェア**: `X-Request-Id` をリクエスト単位で生成・伝搬


## Tech Stack

| Category | Technology |
|----------|-----------|
| Language | Go 1.24 |
| HTTP | net/http (標準ライブラリ) |
| Database | MySQL 8.0 (sqlx) |
| Object Storage | S3 互換 (MinIO / aws-sdk-go-v2) |
| Serialization | Apache Parquet (apache/arrow-go) |
| Metrics | Prometheus + Grafana + Pushgateway |
| Security | bcrypt (golang.org/x/crypto), SHA-256 |
| Container | Docker / Docker Compose |

## Getting Started

### Prerequisites

- Go 1.24+
- Docker / Docker Compose

### Setup

TODO

## Testing

180 テスト関数、15 パッケージでテスト実行。

```
internal/domain/task/         ... 値オブジェクト・エンティティ・ファクトリ
internal/domain/user/         ... 値オブジェクト・エンティティ
internal/usecase/task/create/ ... ユースケース (mock runner + mock repo)
internal/usecase/task/update/ ... ユースケース
internal/usecase/task/delete/ ... ユースケース
internal/usecase/task/list/   ... カーソルエンコード/デコード
internal/usecase/user/register/ ... ユースケース
internal/presentation/http/handler/task/ ... HTTP ハンドラー (httptest)
internal/presentation/http/handler/user/ ... HTTP ハンドラー
internal/presentation/http/middleware/   ... RequestID, AccessLog
internal/presentation/http/responder/    ... JSON レスポンダー
internal/worker/outbox/       ... Outbox Worker
internal/worker/compaction/   ... Compaction Worker
internal/infra/repository/    ... リポジトリ
```

テスト方針:
- 外部モック不要: 手書き mock / stub のみ。testify 等の外部ライブラリ不使用
- テーブル駆動テスト + `t.Parallel()` で並列実行
- Handler テストは `httptest.NewRecorder` + 実ユースケース (mock 注入) で HTTP レベル検証

## Deploy

> TODO

## Database Schema

```
users                        task                         task_events
+-----------------------+    +-----------------------+    +-----------------------+
| id          CHAR(36)  |    | id          CHAR(36)  |    | id          CHAR(36)  |
| email       VARCHAR   |    | user_id     CHAR(36)  |--->| user_id     CHAR(36)  |
| password_hash VARCHAR |    | title       VARCHAR   |    | task_id     CHAR(36)  |
| user_name   VARCHAR   |    | description VARCHAR   |    | request_id  CHAR(36)  |
| email_verified_at     |    | status      VARCHAR   |    | event_type  VARCHAR   |
| version     BIGINT    |    | due_date    DATETIME  |    | occurred_at DATETIME  |
+-----------------------+    | version     BIGINT    |    | emitted_at  DATETIME  |
         |                   +-----------------------+    | payload     JSON      |
         v                            |                   | attempt_count INT     |
email_verification_tokens             |                   | lease_owner VARCHAR   |
+-----------------------+             |                   | lease_until DATETIME  |
| id          CHAR(36)  |             v                   +-----------------------+
| user_id     CHAR(36)  |    task_events_dlq                       |
| token_hash  CHAR(64)  |    +-----------------------+             v
| expires_at  DATETIME  |    | id          CHAR(36)  |    S3 (MinIO)
| used_at     DATETIME  |    | last_error  TEXT      |    +-----------------------+
+-----------------------+    | dead_at     DATETIME  |    | raw/   ... JSONL      |
                             +-----------------------+    | compacted/ ... Parquet |
                                                          +-----------------------+
```

## License

[MIT](LICENSE)
