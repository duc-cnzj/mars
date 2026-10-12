CREATE TABLE `project_hibernates` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `enabled` bool NOT NULL DEFAULT 0,
  `state` varchar(32) NOT NULL DEFAULT "awake",
  `snapshot` json NULL,
  `project_id` bigint NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `projecthibernate_project_id` (`project_id`),
  CONSTRAINT `project_hibernates_projects_hibernate` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL
) CHARSET utf8mb4 COLLATE utf8mb4_bin;

-- 项目定时休眠（夜间缩容）运行态表：每项目一行（唯一索引 project_id），存项目级开关、
-- 运行态与休眠前副本字段快照。副本字段映射与时间窗是仓库级配置（repos.mars_config.hibernate）。
--
-- ⚠️ 与既有迁移一致：本文件是面向外部 provisioning 的产物，应用不会自动执行；应用启动在
-- `db_auto_migrate: true` 时走 ent 原生迁移（读 ent/migrate/schema.go），与本目录无关。
-- 上线顺序仍应先执行本迁移、再发新版本（缺表会让休眠对账查询报 1146）。
