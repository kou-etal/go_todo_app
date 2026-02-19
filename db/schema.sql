CREATE TABLE `users`
(
    `id`                CHAR(36)        NOT NULL COMMENT 'ユーザーID（UUID）',
    `email`             VARCHAR(254)    NOT NULL COMMENT 'メールアドレス',
    `password_hash`     VARCHAR(80)     NOT NULL COMMENT 'パスワードハッシュ',
    `user_name`         VARCHAR(20)     NOT NULL COMMENT 'ユーザー名',
    `email_verified_at` DATETIME(6)     NULL COMMENT 'メール認証日時',
    `created_at`        DATETIME(6)     NOT NULL COMMENT '作成日時',
    `updated_at`        DATETIME(6)     NOT NULL COMMENT '更新日時',
    `version`           BIGINT UNSIGNED NOT NULL COMMENT '楽観ロック用バージョン',

    PRIMARY KEY (`id`),
    UNIQUE KEY `uix_users_email` (`email`),
    UNIQUE KEY `uix_users_user_name` (`user_name`)
)
ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='ユーザー';

CREATE TABLE `task`
(
    `id`          CHAR(36) NOT NULL COMMENT 'タスクの識別子',
    `user_id`     CHAR(36) NOT NULL COMMENT '所有者ユーザーID(UUID)',
    `title`       VARCHAR(20) NOT NULL COMMENT 'タスクのタイトル',
    `description` VARCHAR(1000) NOT NULL COMMENT 'タスクの説明',
    `status`      VARCHAR(20)  NOT NULL COMMENT 'タスクの状態',
    `due_date`    DATETIME(6) NOT NULL COMMENT 'deadline',
    `created_at`  DATETIME(6) NOT NULL COMMENT 'レコード作成日時',
    `updated_at`  DATETIME(6) NOT NULL COMMENT 'レコード修正日時',
    `version`     BIGINT UNSIGNED NOT NULL COMMENT 'バージョン',
    `due_is_null` TINYINT(1)
    AS (due_date IS NULL) STORED
    NOT NULL COMMENT '現状due_date NOT NULLなので常に0',
    PRIMARY KEY (`id`),
    KEY `idx_task_user_created_sort` (`user_id`, `created_at`, `id`),
    KEY `idx_task_user_due_sort` (`user_id`, `due_is_null`, `due_date`, `id`),
    CONSTRAINT `fk_task_user_id`
      FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
      ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='タスク';

CREATE TABLE `email_verification_tokens` (
  `id`         CHAR(36) NOT NULL COMMENT 'トークン行の識別子',
  `user_id`    CHAR(36) NOT NULL COMMENT '対象ユーザーID(UUID)',
  `token_hash` CHAR(64) NOT NULL COMMENT 'SHA-256(token) を hex 化したもの',
  `expires_at` DATETIME(6) NOT NULL COMMENT '有効期限',
  `used_at`    DATETIME(6) NULL COMMENT '使用日時(未使用ならNULL)',
  `created_at` DATETIME(6) NOT NULL COMMENT '作成日時',

  PRIMARY KEY (`id`),

  UNIQUE KEY `uix_token_hash` (`token_hash`) USING BTREE,
  KEY `idx_user_id_created_at` (`user_id`, `created_at`) USING BTREE,
  KEY `idx_expires_at` (`expires_at`) USING BTREE,

  CONSTRAINT `fk_email_verification_tokens_user_id`
    FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='メール認証トークン';

CREATE TABLE `task_events` (
  `id`               CHAR(36)        NOT NULL COMMENT 'イベントID（UUID）',
  `user_id`          CHAR(36)        NOT NULL COMMENT '操作ユーザーID',
  `task_id`          CHAR(36)        NOT NULL COMMENT '対象タスクID',
  `request_id`       CHAR(36)        NOT NULL COMMENT 'リクエストID（トレーシング用）',
  `event_type`       VARCHAR(20)     NOT NULL COMMENT 'イベント種別（created/updated/deleted）',
  `occurred_at`      DATETIME(6)     NOT NULL COMMENT 'イベント発生日時',
  `emitted_at`       DATETIME(6)     NULL     COMMENT '外部配信日時（未配信ならNULL）',
  `schema_version`   INT UNSIGNED    NOT NULL COMMENT 'ペイロードのスキーマバージョン',
  `payload`          JSON            NOT NULL COMMENT 'イベントペイロード',
  `next_attempt_at`  DATETIME(6)     NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
    COMMENT 'リトライ可能時刻。初回は即時',
  `attempt_count`    INT UNSIGNED    NOT NULL DEFAULT 0
    COMMENT '試行回数',
  `lease_owner`      VARCHAR(255)    NULL
    COMMENT 'リースを保持するワーカーID',
  `lease_until`      DATETIME(6)     NULL
    COMMENT 'リース有効期限',
  `claimed_at`       DATETIME(6)     NULL
    COMMENT 'claim した時刻',

  PRIMARY KEY (`id`),
  KEY `idx_task_events_emitted` (`emitted_at`),
  KEY `idx_task_events_claimable` (`emitted_at`, `next_attempt_at`, `lease_until`),
  CONSTRAINT `fk_task_events_task_id`
    FOREIGN KEY (`task_id`) REFERENCES `task` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='タスクイベント（Outbox）';

CREATE TABLE `task_events_dlq` (
  `id`             CHAR(36)        NOT NULL COMMENT '元イベントID',
  `user_id`        CHAR(36)        NOT NULL COMMENT '操作ユーザーID',
  `task_id`        CHAR(36)        NOT NULL COMMENT '対象タスクID',
  `request_id`     CHAR(36)        NOT NULL COMMENT 'リクエストID（トレーシング用）',
  `event_type`     VARCHAR(20)     NOT NULL COMMENT 'イベント種別（created/updated/deleted）',
  `occurred_at`    DATETIME(6)     NOT NULL COMMENT 'イベント発生日時',
  `schema_version` INT UNSIGNED    NOT NULL COMMENT 'ペイロードのスキーマバージョン',
  `payload`        JSON            NOT NULL COMMENT 'イベントペイロード',
  `attempt_count`  INT UNSIGNED    NOT NULL COMMENT '最終試行回数',
  `last_error`     TEXT            NULL     COMMENT '最後のエラーメッセージ',
  `dead_at`        DATETIME(6)     NOT NULL COMMENT 'DLQ 投入日時',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配信失敗イベント（DLQ）';
