CREATE DATABASE IF NOT EXISTS `skyvv`
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE `skyvv`;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

CREATE TABLE IF NOT EXISTS `sys_dept` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `parent_id` BIGINT NOT NULL DEFAULT 0,
  `ancestors` VARCHAR(512) NOT NULL DEFAULT '',
  `name` VARCHAR(128) NOT NULL,
  `leader` VARCHAR(64) NOT NULL DEFAULT '',
  `phone` VARCHAR(32) NOT NULL DEFAULT '',
  `email` VARCHAR(128) NOT NULL DEFAULT '',
  `sort` INT NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1,
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `updated_by` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  `remark` VARCHAR(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_sys_dept_parent_id` (`parent_id`),
  KEY `idx_sys_dept_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='部门表';

CREATE TABLE IF NOT EXISTS `sys_role` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(64) NOT NULL,
  `role_key` VARCHAR(64) NOT NULL,
  `data_scope` VARCHAR(32) NOT NULL DEFAULT 'custom',
  `status` TINYINT NOT NULL DEFAULT 1,
  `sort` INT NOT NULL DEFAULT 0,
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `updated_by` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  `remark` VARCHAR(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_role_key` (`role_key`),
  UNIQUE KEY `uk_sys_role_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

CREATE TABLE IF NOT EXISTS `sys_menu` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `parent_id` BIGINT NOT NULL DEFAULT 0,
  `name` VARCHAR(128) NOT NULL,
  `menu_type` VARCHAR(16) NOT NULL COMMENT 'directory/menu/button',
  `path` VARCHAR(255) NOT NULL DEFAULT '',
  `component` VARCHAR(255) NOT NULL DEFAULT '',
  `route_name` VARCHAR(128) NOT NULL DEFAULT '',
  `perms` VARCHAR(255) NOT NULL DEFAULT '',
  `icon` VARCHAR(128) NOT NULL DEFAULT '',
  `visible` TINYINT NOT NULL DEFAULT 1,
  `status` TINYINT NOT NULL DEFAULT 1,
  `sort` INT NOT NULL DEFAULT 0,
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `updated_by` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  `remark` VARCHAR(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_sys_menu_parent_id` (`parent_id`),
  KEY `idx_sys_menu_type` (`menu_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单表';

CREATE TABLE IF NOT EXISTS `sys_user` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `dept_id` BIGINT NOT NULL DEFAULT 0,
  `username` VARCHAR(64) NOT NULL,
  `nickname` VARCHAR(64) NOT NULL DEFAULT '',
  `real_name` VARCHAR(64) NOT NULL DEFAULT '',
  `phone` VARCHAR(32) NOT NULL DEFAULT '',
  `email` VARCHAR(128) NOT NULL DEFAULT '',
  `password_hash` VARCHAR(255) NOT NULL DEFAULT '',
  `avatar` VARCHAR(255) NOT NULL DEFAULT '',
  `status` TINYINT NOT NULL DEFAULT 1,
  `last_login_ip` VARCHAR(64) NOT NULL DEFAULT '',
  `last_login_at` DATETIME NULL DEFAULT NULL,
  `password_reset_required` TINYINT NOT NULL DEFAULT 1,
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `updated_by` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  `remark` VARCHAR(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_user_username` (`username`),
  UNIQUE KEY `uk_sys_user_phone` (`phone`),
  KEY `idx_sys_user_dept_id` (`dept_id`),
  KEY `idx_sys_user_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

CREATE TABLE IF NOT EXISTS `sys_user_role` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL,
  `role_id` BIGINT NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_user_role` (`user_id`, `role_id`),
  KEY `idx_sys_user_role_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

CREATE TABLE IF NOT EXISTS `sys_role_menu` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `role_id` BIGINT NOT NULL,
  `menu_id` BIGINT NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_role_menu` (`role_id`, `menu_id`),
  KEY `idx_sys_role_menu_menu_id` (`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色菜单关联表';

CREATE TABLE IF NOT EXISTS `sys_config` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `config_group` VARCHAR(64) NOT NULL DEFAULT 'default',
  `config_key` VARCHAR(128) NOT NULL,
  `config_value` LONGTEXT NOT NULL,
  `value_type` VARCHAR(32) NOT NULL DEFAULT 'string',
  `is_encrypted` TINYINT NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1,
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `updated_by` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  `remark` VARCHAR(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_config_group_key` (`config_group`, `config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

CREATE TABLE IF NOT EXISTS `sys_login_log` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL DEFAULT 0,
  `username` VARCHAR(64) NOT NULL DEFAULT '',
  `login_ip` VARCHAR(64) NOT NULL DEFAULT '',
  `user_agent` VARCHAR(512) NOT NULL DEFAULT '',
  `login_status` TINYINT NOT NULL DEFAULT 1,
  `message` VARCHAR(500) NOT NULL DEFAULT '',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_sys_login_log_user_id` (`user_id`),
  KEY `idx_sys_login_log_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='登录日志表';

CREATE TABLE IF NOT EXISTS `storage` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `type` VARCHAR(16) NOT NULL COMMENT 'local/s3',
  `endpoint` VARCHAR(255) NOT NULL DEFAULT '',
  `access_key` VARCHAR(255) NOT NULL DEFAULT '',
  `secret_key` VARCHAR(255) NOT NULL DEFAULT '',
  `bucket` VARCHAR(128) NOT NULL DEFAULT '',
  `region` VARCHAR(128) NOT NULL DEFAULT '',
  `local_path` VARCHAR(1024) NOT NULL DEFAULT '',
  `quota_bytes` BIGINT NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1,
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `updated_by` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  `remark` VARCHAR(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_storage_name` (`name`),
  KEY `idx_storage_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='存储定义表';

CREATE TABLE IF NOT EXISTS `file_filter` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `filter_scope` VARCHAR(32) NOT NULL COMMENT 'extension/path',
  `list_type` VARCHAR(32) NOT NULL COMMENT 'blacklist/whitelist',
  `patterns` JSON NOT NULL,
  `status` TINYINT NOT NULL DEFAULT 1,
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `updated_by` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  `remark` VARCHAR(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_file_filter_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件过滤规则表';

CREATE TABLE IF NOT EXISTS `sync_regex` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `source_field` VARCHAR(32) NOT NULL COMMENT 'path/filename',
  `regexp` VARCHAR(1024) NOT NULL,
  `as_path` TINYINT NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1,
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `updated_by` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  `remark` VARCHAR(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sync_regex_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='正则提取规则表';

CREATE TABLE IF NOT EXISTS `sync_regex_mapping` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `regex_id` BIGINT NOT NULL,
  `capture_index` INT NOT NULL,
  `tag_key` VARCHAR(128) NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sync_regex_mapping` (`regex_id`, `capture_index`, `tag_key`),
  KEY `idx_sync_regex_mapping_regex_id` (`regex_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='正则提取映射表';

CREATE TABLE IF NOT EXISTS `sync_image_process` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `processor_type` VARCHAR(64) NOT NULL DEFAULT '',
  `config_json` JSON NOT NULL,
  `status` TINYINT NOT NULL DEFAULT 1,
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `updated_by` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  `remark` VARCHAR(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sync_image_process_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图片处理规则表';

CREATE TABLE IF NOT EXISTS `sync_alarm_group` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `receivers` JSON NOT NULL,
  `status` TINYINT NOT NULL DEFAULT 1,
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `updated_by` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  `remark` VARCHAR(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sync_alarm_group_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警组表';

CREATE TABLE IF NOT EXISTS `sync_alarm_rule` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `cpu_threshold` DECIMAL(5,2) NOT NULL DEFAULT 0.00,
  `mem_threshold` DECIMAL(5,2) NOT NULL DEFAULT 0.00,
  `disk_threshold` DECIMAL(5,2) NOT NULL DEFAULT 0.00,
  `cpu_consecutive_times` INT NOT NULL DEFAULT 0,
  `mem_consecutive_times` INT NOT NULL DEFAULT 0,
  `heartbeat_timeout_seconds` INT NOT NULL DEFAULT 0,
  `send_frequency_seconds` INT NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1,
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `updated_by` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  `remark` VARCHAR(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sync_alarm_rule_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警规则表';

CREATE TABLE IF NOT EXISTS `sync_alarm_message` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `channel_type` VARCHAR(32) NOT NULL COMMENT 'email/wecom',
  `config_json` JSON NOT NULL,
  `status` TINYINT NOT NULL DEFAULT 1,
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `updated_by` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  `remark` VARCHAR(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sync_alarm_message_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警消息通道表';

CREATE TABLE IF NOT EXISTS `sync_agent_group` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `storage_id` BIGINT NOT NULL,
  `ip_range` VARCHAR(255) NOT NULL DEFAULT '',
  `path_prefix` VARCHAR(512) NOT NULL DEFAULT '',
  `interval_time` BIGINT NOT NULL DEFAULT 60,
  `del_time_days` BIGINT NOT NULL DEFAULT 0,
  `max_workers` INT NOT NULL DEFAULT 4,
  `work_windows` JSON NOT NULL,
  `file_filter_id` BIGINT NOT NULL DEFAULT 0,
  `regex_id` BIGINT NOT NULL DEFAULT 0,
  `image_process_id` BIGINT NOT NULL DEFAULT 0,
  `alarm_group_id` BIGINT NOT NULL DEFAULT 0,
  `log_enabled` TINYINT NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1,
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `updated_by` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  `remark` VARCHAR(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sync_agent_group_name` (`name`),
  KEY `idx_sync_agent_group_storage_id` (`storage_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Agent分组策略表';

CREATE TABLE IF NOT EXISTS `sync_agent` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `host_sn` VARCHAR(128) NOT NULL,
  `host_name` VARCHAR(128) NOT NULL DEFAULT '',
  `ip` VARCHAR(64) NOT NULL,
  `group_id` BIGINT NOT NULL DEFAULT 0,
  `storage_id` BIGINT NOT NULL DEFAULT 0,
  `source_paths` JSON NULL,
  `storage_metrics` JSON NULL,
  `path_prefix` VARCHAR(512) NOT NULL DEFAULT '',
  `version` VARCHAR(64) NOT NULL DEFAULT '',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '0=offline,1=online,2=disabled',
  `tags` JSON NOT NULL,
  `last_access_time` DATETIME NULL DEFAULT NULL,
  `last_commit_time` DATETIME NULL DEFAULT NULL,
  `last_heartbeat_raw` JSON NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  `remark` VARCHAR(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sync_agent_host_sn` (`host_sn`),
  UNIQUE KEY `uk_sync_agent_ip` (`ip`),
  KEY `idx_sync_agent_group_id` (`group_id`),
  KEY `idx_sync_agent_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Agent主表';

CREATE TABLE IF NOT EXISTS `sync_agent_monitor` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `agent_id` BIGINT NOT NULL,
  `cpu_usage` DECIMAL(6,2) NOT NULL DEFAULT 0.00,
  `mem_usage` DECIMAL(6,2) NOT NULL DEFAULT 0.00,
  `disk_usage_json` JSON NOT NULL,
  `heartbeat_at` DATETIME NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sync_agent_monitor_agent_id` (`agent_id`),
  KEY `idx_sync_agent_monitor_heartbeat_at` (`heartbeat_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Agent最新监控快照表';

CREATE TABLE IF NOT EXISTS `sync_agent_monitor_series` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `agent_id` BIGINT NOT NULL,
  `metric_time` DATETIME NOT NULL,
  `cpu_usage` DECIMAL(6,2) NOT NULL DEFAULT 0.00,
  `mem_usage` DECIMAL(6,2) NOT NULL DEFAULT 0.00,
  `disk_usage_json` JSON NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_sync_agent_monitor_series_agent_time` (`agent_id`, `metric_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Agent监控趋势表';

CREATE TABLE IF NOT EXISTS `sync_agent_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `agent_id` BIGINT NOT NULL,
  `group_id` BIGINT NOT NULL DEFAULT 0,
  `task_type` VARCHAR(16) NOT NULL COMMENT 'sync/scan',
  `source_path` VARCHAR(1024) NOT NULL DEFAULT '',
  `task_start_time` DATETIME NULL DEFAULT NULL,
  `file_count` BIGINT NOT NULL DEFAULT 0,
  `file_size_bytes` BIGINT NOT NULL DEFAULT 0,
  `error_count` BIGINT NOT NULL DEFAULT 0,
  `tar_list_path` VARCHAR(1024) NOT NULL DEFAULT '',
  `status` VARCHAR(32) NOT NULL DEFAULT 'success',
  `message` VARCHAR(500) NOT NULL DEFAULT '',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_sync_agent_logs_agent_id` (`agent_id`),
  KEY `idx_sync_agent_logs_task_type` (`task_type`),
  KEY `idx_sync_agent_logs_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Agent任务日志表';

CREATE TABLE IF NOT EXISTS `sync_scan_report` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `agent_id` BIGINT NOT NULL,
  `group_id` BIGINT NOT NULL DEFAULT 0,
  `source_path` VARCHAR(1024) NOT NULL DEFAULT '',
  `task_start_time` DATETIME NULL DEFAULT NULL,
  `total_files` BIGINT NOT NULL DEFAULT 0,
  `total_size_bytes` BIGINT NOT NULL DEFAULT 0,
  `report_json` JSON NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_sync_scan_report_agent_id` (`agent_id`),
  KEY `idx_sync_scan_report_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='扫描报告表';

CREATE TABLE IF NOT EXISTS `sync_scan_report_node` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `report_id` BIGINT NOT NULL,
  `path` VARCHAR(1024) NOT NULL,
  `level` INT NOT NULL DEFAULT 0,
  `total_files` BIGINT NOT NULL DEFAULT 0,
  `total_size_bytes` BIGINT NOT NULL DEFAULT 0,
  `type_summary_json` JSON NOT NULL,
  `size_summary_json` JSON NOT NULL,
  `period_summary_json` JSON NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_sync_scan_report_node_report_id` (`report_id`),
  KEY `idx_sync_scan_report_node_path` (`path`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='扫描报告目录节点表';

CREATE TABLE IF NOT EXISTS `sync_agent_statistic` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `stat_date` DATE NOT NULL,
  `group_id` BIGINT NOT NULL DEFAULT 0,
  `agent_count` BIGINT NOT NULL DEFAULT 0,
  `online_count` BIGINT NOT NULL DEFAULT 0,
  `file_count` BIGINT NOT NULL DEFAULT 0,
  `file_size_bytes` BIGINT NOT NULL DEFAULT 0,
  `error_count` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sync_agent_statistic` (`stat_date`, `group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Agent日统计表';

CREATE TABLE IF NOT EXISTS `sync_file_log` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL DEFAULT 0,
  `username` VARCHAR(64) NOT NULL DEFAULT '',
  `storage_id` BIGINT NOT NULL DEFAULT 0,
  `storage_name` VARCHAR(128) NOT NULL DEFAULT '',
  `file_path` VARCHAR(2048) NOT NULL,
  `operation_type` VARCHAR(32) NOT NULL,
  `result_status` VARCHAR(32) NOT NULL DEFAULT 'success',
  `client_ip` VARCHAR(64) NOT NULL DEFAULT '',
  `message` VARCHAR(500) NOT NULL DEFAULT '',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_sync_file_log_user_id` (`user_id`),
  KEY `idx_sync_file_log_storage_id` (`storage_id`),
  KEY `idx_sync_file_log_operation_type` (`operation_type`),
  KEY `idx_sync_file_log_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件操作审计日志表';

CREATE TABLE IF NOT EXISTS `sync_file_progress` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `task_type` VARCHAR(64) NOT NULL,
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending',
  `total_count` BIGINT NOT NULL DEFAULT 0,
  `success_count` BIGINT NOT NULL DEFAULT 0,
  `failed_count` BIGINT NOT NULL DEFAULT 0,
  `payload_json` JSON NOT NULL,
  `result_json` JSON NULL,
  `started_at` DATETIME NULL DEFAULT NULL,
  `finished_at` DATETIME NULL DEFAULT NULL,
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `remark` VARCHAR(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_sync_file_progress_status` (`status`),
  KEY `idx_sync_file_progress_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='异步任务进度表';

CREATE TABLE IF NOT EXISTS `sync_alarm_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `agent_id` BIGINT NOT NULL,
  `group_id` BIGINT NOT NULL DEFAULT 0,
  `rule_id` BIGINT NOT NULL DEFAULT 0,
  `message_channel_id` BIGINT NOT NULL DEFAULT 0,
  `alert_level` VARCHAR(32) NOT NULL DEFAULT 'warning',
  `alert_title` VARCHAR(255) NOT NULL DEFAULT '',
  `alert_body` TEXT NOT NULL,
  `send_status` VARCHAR(32) NOT NULL DEFAULT 'pending',
  `failure_reason` VARCHAR(500) NOT NULL DEFAULT '',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_sync_alarm_logs_agent_id` (`agent_id`),
  KEY `idx_sync_alarm_logs_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警日志表';

CREATE TABLE IF NOT EXISTS `sync_agent_version` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `version` VARCHAR(64) NOT NULL,
  `filename` VARCHAR(255) NOT NULL,
  `file_path` VARCHAR(1024) NOT NULL DEFAULT '',
  `md5` VARCHAR(64) NOT NULL DEFAULT '',
  `file_size_bytes` BIGINT NOT NULL DEFAULT 0,
  `is_current` TINYINT NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1,
  `release_note` TEXT NULL,
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `updated_by` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sync_agent_version_version` (`version`),
  KEY `idx_sync_agent_version_current` (`is_current`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Agent版本表';

CREATE TABLE IF NOT EXISTS `sync_license` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `license_code` LONGTEXT NOT NULL,
  `serial_number` VARCHAR(255) NOT NULL DEFAULT '',
  `max_agent_count` INT NOT NULL DEFAULT 0,
  `issued_at` DATETIME NULL DEFAULT NULL,
  `expired_at` DATETIME NULL DEFAULT NULL,
  `trial_days` INT NOT NULL DEFAULT 30,
  `status` TINYINT NOT NULL DEFAULT 1,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `remark` VARCHAR(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='授权信息表';

CREATE TABLE IF NOT EXISTS `sync_user_auth` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL,
  `storage_id` BIGINT NOT NULL,
  `can_view` TINYINT NOT NULL DEFAULT 1,
  `can_upload` TINYINT NOT NULL DEFAULT 0,
  `can_download` TINYINT NOT NULL DEFAULT 0,
  `can_delete` TINYINT NOT NULL DEFAULT 0,
  `can_batch_download` TINYINT NOT NULL DEFAULT 0,
  `can_download_to_server` TINYINT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sync_user_auth` (`user_id`, `storage_id`),
  KEY `idx_sync_user_auth_storage_id` (`storage_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户存储权限表';

INSERT INTO `sys_dept` (`id`, `parent_id`, `ancestors`, `name`, `leader`, `status`, `remark`)
VALUES (1, 0, '0', 'SkyBase', 'system', 1, '默认根部门')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`), `remark` = VALUES(`remark`);

INSERT INTO `sys_role` (`id`, `name`, `role_key`, `data_scope`, `status`, `remark`)
VALUES (1, '超级管理员', 'super_admin', 'all', 1, '默认超级管理员角色')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`), `data_scope` = VALUES(`data_scope`);

INSERT INTO `sys_user` (`id`, `dept_id`, `username`, `nickname`, `real_name`, `password_hash`, `status`, `password_reset_required`, `remark`)
VALUES (1, 1, 'admin', '管理员', 'SkyBase Admin', 'INIT_PENDING', 1, 1, '默认管理员账号，首次接入认证时请重置密码')
ON DUPLICATE KEY UPDATE `nickname` = VALUES(`nickname`), `remark` = VALUES(`remark`);

INSERT INTO `sys_user_role` (`user_id`, `role_id`)
VALUES (1, 1)
ON DUPLICATE KEY UPDATE `role_id` = VALUES(`role_id`);

INSERT INTO `sys_config` (`config_group`, `config_key`, `config_value`, `value_type`, `is_encrypted`, `status`, `remark`)
VALUES
  ('system', 'app.name', 'SkyBase', 'string', 0, 1, '应用名称'),
  ('system', 'app.env', 'dev', 'string', 0, 1, '运行环境'),
  ('license', 'trial.enabled', 'true', 'bool', 0, 1, '是否允许试用')
ON DUPLICATE KEY UPDATE `config_value` = VALUES(`config_value`), `remark` = VALUES(`remark`);

SET FOREIGN_KEY_CHECKS = 1;
