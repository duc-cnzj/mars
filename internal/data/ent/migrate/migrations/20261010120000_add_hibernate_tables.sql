CREATE TABLE `project_hibernates` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `enabled` bool NOT NULL DEFAULT 0,
  `state` varchar(32) NOT NULL DEFAULT "awake",
  `snapshot` json NULL,
  `window_names` json NULL,
  `project_id` bigint NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `project_id` (`project_id`),
  UNIQUE INDEX `projecthibernate_project_id` (`project_id`),
  CONSTRAINT `project_hibernates_projects_hibernate` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL
) CHARSET utf8mb4 COLLATE utf8mb4_bin;

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

-- 定时休眠（夜间缩容）功能建表：全局时间窗字典 + 项目运行态（1:1）。
--
-- project_hibernates：每项目一行（唯一索引 project_id），存项目级开关、运行态（awake/sleeping）
-- 与休眠前副本字段快照。副本字段映射与时间窗是仓库级配置（repos.mars_config.hibernate）；
-- window_names 为项目选定的时间窗名子集（引用仓库可用窗列表的子集）。
--
-- hibernate_windows：全局共享的命名时段（name/start/end），各仓库与项目按名字引用其中若干。
-- 按名字引用（非 ID）是因为 mars_config 是 JSON blob、会经导出/导入跨环境往返，用 ID 会在导入后
-- 错位。description 定为 varchar(255)：与 ent schema（field.String 无 MaxLen → 默认 255）对齐。
--
-- project_id 上的两个唯一键（project_id + projecthibernate_project_id）与 ent 原生迁移产出
-- 逐字节对齐：schema 里「1:1 唯一边」与「显式 index.Fields("project_id").Unique()」各生成一个
-- 唯一键，本文件须照抄，否则外部 provisioning 结果与运行期（ent auto-migrate）分歧。
--
-- ⚠️ 与既有迁移一致：本文件是面向外部 provisioning 的产物，应用不会自动执行；应用启动在
-- `db_auto_migrate: true` 时走 ent 原生迁移（读 ent/migrate/schema.go），与本目录无关。
-- 上线顺序仍应先执行本迁移、再发新版本（缺表会让休眠对账/时间窗查询报 1146）。
