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
--TODO:CHAR(36)はインデックス重い。CHAR(26)ULIDやBINARY(16)も検討
CREATE TABLE `task`
(
    `id`          CHAR(36) NOT NULL COMMENT 'タスクの識別子',
    `title`       VARCHAR(20) NOT NULL COMMENT 'タスクのタイトル',
    `description` VARCHAR(1000) NOT NULL COMMENT 'タスクの説明',
    `status`      VARCHAR(20)  NOT NULL COMMENT 'タスクの状態',
    `due_dte`     DATETIME(6) NOT NULL COMMENT 'deadline',
    `created_at`     DATETIME(6) NOT NULL COMMENT 'レコード作成日時',
    `updated_at`     DATETIME(6) NOT NULL COMMENT 'レコード修正日時',
    `version`     BIGINT UNSIGNED NOT NULL COMMENT 'バージョン',
    --負債
    `due_is_null` TINYINT(1)
    AS (dueDate IS NULL) STORED
    NOT NULL COMMENT '現状due_date NOT NULLなので常に0',
    PRIMARY KEY (`id`)
    --indexないと全件とってからソートするから重くなる
    KEY `idx_task_created_sort` (`created_at`, `id`),
    KEY `idx_task_due_sort` (`due_is_null`, `due_date`, `id`)
) Engine=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='タスク';
--duedate not nullやのに仕様変更のためにdue_is_nullっていう主張は不可