CREATE DATABASE IF NOT EXISTS `ServiceA`;

CREATE TABLE IF NOT EXISTS `t_user`(
    `id` varchar(36) NOT NULL,
    `name` varchar(32) NOT NULL,
    `password` varchar(128) NOT NULL,
    `create_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `last_login_at` TIMESTAMP DEFAULT NULL,
    PRIMARY KEY(`id`),
    UNIQUE KEY(`name`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;