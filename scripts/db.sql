CREATE DATABASE IF NOT EXISTS `IAMService` CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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

CREATE TABLE IF NOT EXISTS `t_jwt_keys` (
  `id` varchar(36) NOT NULL,
  `data` text NOT NULL,
  `sid` varchar(36) NOT NULL COMMENT 'set id', 
  `created_at` timestamp DEFAULT current_timestamp,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_sid` (`sid`),
  KEY `idx_created_at` (`created_at`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `t_jwt_blacklist` (
  `id` varchar(36) NOT NULL,
  `created_at` timestamp DEFAULT current_timestamp,
  PRIMARY KEY (`id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;