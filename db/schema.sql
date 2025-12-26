CREATE TABLE `user`
(
    `id`       BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ユーザーの識別子',
    `name`     varchar(20) NOT NULL COMMENT 'ユーザー名',
    `password` VARCHAR(80) NOT NULL COMMENT 'パスワードハッシュ',
    `role`     VARCHAR(80) NOT NULL COMMENT 'ロール',
    `created`  DATETIME(6) NOT NULL COMMENT 'レコード作成日時',
    `modified` DATETIME(6) NOT NULL COMMENT 'レコード修正日時',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uix_name` (`name`) USING BTREE
) Engine=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='ユーザー';
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