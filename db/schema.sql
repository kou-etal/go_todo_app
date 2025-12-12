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
    `dueDate`     DATETIME(6) NOT NULL COMMENT 'deadline',
    `created`     DATETIME(6) NOT NULL COMMENT 'レコード作成日時',
    `updated`     DATETIME(6) NOT NULL COMMENT 'レコード修正日時',
    `version`     BIGINT UNSIGNED NOT NULL COMMENT 'バージョン',
    PRIMARY KEY (`id`)
) Engine=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='タスク';