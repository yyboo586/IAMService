CREATE DATABASE IF NOT EXISTS `ServiceA` CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `t_user`(
    `id` varchar(36) NOT NULL,
    `name` varchar(32) NOT NULL,
    `password` varchar(128) NOT NULL,
    `email` varchar(128) DEFAULT NULL,
    `create_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `last_login_at` TIMESTAMP DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_name` (`name`),
    KEY `idx_last_login_at` (`last_login_at`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户信息';
