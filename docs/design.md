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

## 3.アーキテクチャ
DDDを採用
- cmd: エントリポイント（サーバー起動）
- handler(presentation): HTTP層（リクエスト/レスポンスの整形）
- usecase: ユースケース実行ロジック（アプリケーションサービス）
- domain: エンティティ、Value Object、ビジネスロジックの実装
- infra: 外部I/O（DB・外部API等）の具体実装

依存関係のルール
- domain は他レイヤーに依存しない
- handler は usecase に依存する
- usecase は domain に依存するが、infra/handler には依存しない
- infra は domain に依存する
- cmd は全レイヤーを束ねるが、ビジネスロジックは持たない

*ここにinterfaceをどに置くか後ほど詳しく記述する*


TODO:トランザクション、metrics、perf、公開API、
middlewareは例えばuseridを与えたいってなった時にctxを通す
