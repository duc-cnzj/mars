CREATE TABLE `hibernate_windows` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `name` varchar(64) NOT NULL,
  `start` varchar(5) NOT NULL,
  `end` varchar(5) NOT NULL,
  `description` varchar(255) NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `hibernatewindow_name` (`name`)
) CHARSET utf8mb4 COLLATE utf8mb4_bin;

-- 全局共享的休眠时间窗表：管理员在后台维护若干命名时段（name/start/end），
-- 各仓库的定时休眠配置（repos.mars_config.hibernate.window_names）按名字引用其中若干。
-- 仓库按名字引用（非 ID）——mars_config 是 JSON blob，会经导出/导入跨环境往返，
-- 用 ID 会在导入后错位；name 在当前环境内唯一即可。
--
-- description 定为 varchar(255)：与 ent schema（field.String 无 MaxLen → 默认 255）逐字节对齐，
-- 早期手写的 longtext 会让 ent-diff 持续报漂移。
--
-- ⚠️ 与既有迁移一致：本文件是面向外部 provisioning 的产物，应用不会自动执行；应用启动在
-- `db_auto_migrate: true` 时走 ent 原生迁移（读 ent/migrate/schema.go），与本目录无关。
-- 上线顺序仍应先执行本迁移、再发新版本（缺表会让后台时间窗管理与休眠对账解析报 1146）。
