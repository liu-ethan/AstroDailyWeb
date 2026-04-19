DROP TABLE IF EXISTS `users`;
DROP TABLE IF EXISTS `fortunes`;
DROP TABLE IF EXISTS `verification_codes`;
DROP TABLE IF EXISTS `user_profiles`;

CREATE TABLE `users` (
                         `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                         `email` varchar(255) NOT NULL COMMENT '用户邮箱，登录账号',
                         `password` varchar(255) NOT NULL COMMENT '密码（按需求明文存储）',
                         `is_subscribed` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否订阅每日运势：0-未订阅，1-已订阅',
                         `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
                         `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '信息更新时间',
                         PRIMARY KEY (`id`),
                         UNIQUE KEY `uk_email` (`email`) COMMENT '确保邮箱唯一'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

CREATE TABLE `fortunes` (
                            `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                            `user_id` bigint(20) unsigned NOT NULL COMMENT '关联的用户ID',
                            `target_date` date NOT NULL COMMENT '运势所属的日期 (例如: 2026-04-19)',
                            `content` text NOT NULL COMMENT '大模型生成的运势内容',
                            `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '生成时间',
                            PRIMARY KEY (`id`),
                            UNIQUE KEY `uk_user_date` (`user_id`,`target_date`) COMMENT '联合唯一索引：确保单用户每日只有一条记录',
                            KEY `idx_target_date` (`target_date`) COMMENT '普通索引：用于加速定时任务删除7天前数据的查询'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='每日运势数据表';


CREATE TABLE `verification_codes` (
                                      `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                                      `email` varchar(255) NOT NULL COMMENT '接收验证码的邮箱',
                                      `code` varchar(10) NOT NULL COMMENT '发送的验证码',
                                      `business_type` tinyint(4) NOT NULL COMMENT '业务场景：1-注册，2-重置密码',
                                      `expires_at` timestamp NOT NULL COMMENT '验证码过期时间 (例如: 设定为发送后5分钟)',
                                      `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发送时间',
                                      PRIMARY KEY (`id`),
                                      KEY `idx_email_type` (`email`,`business_type`) COMMENT '查询索引：加速验证时的匹配'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮箱验证码记录表';

CREATE TABLE `user_profiles` (
                                   `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                                   `user_id` bigint(20) unsigned NOT NULL COMMENT '关联用户ID',
                                   `birthday` date NOT NULL COMMENT '出生日期',
                                   `constellation` varchar(32) NOT NULL COMMENT '星座',
                                   `gender` varchar(16) NOT NULL COMMENT '性别',
                                   `city` varchar(64) NOT NULL COMMENT '所在城市',
                                   `occupation` varchar(64) NOT NULL COMMENT '职业',
                                   `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                   `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                   PRIMARY KEY (`id`),
                                   UNIQUE KEY `uk_user_id` (`user_id`) COMMENT '单用户一份资料',
                                   KEY `idx_constellation` (`constellation`) COMMENT '星座索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户运势资料表';
