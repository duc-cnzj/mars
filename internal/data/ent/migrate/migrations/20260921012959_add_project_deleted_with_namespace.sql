-- Modify "projects" table
ALTER TABLE `projects` ADD COLUMN `deleted_with_namespace` bool NOT NULL DEFAULT 0 AFTER `git_commit_date`;

-- 批次标识列：true ⟺ 该项目正处于「随所属空间级联软删」的状态（不变式见
-- internal/data/namespace.go 的 Delete / RestoreDeleted）。存量行一律取列默认值 0。
--
-- 刻意**不回填**本列上线前已软删的项目：那些删除记录已无保留价值。代价是升级前删除的空间
-- 恢复后只还原空间骨架，其下项目不连带还原，需用项目级恢复接口逐个救回（数据不丢，仅不连带）。
--
-- ⚠️ 本文件**不会**被应用自动执行：应用启动仅在 `db_auto_migrate: true` 时调用
-- data.Migrate() → ent Schema.Create —— 那是 ent 原生迁移（读 ent/migrate/schema.go），与本目录
-- 无关；本目录是面向外部 provisioning 的产物，全仓没有任何 `atlas migrate apply` 入口。
-- 而 `db_auto_migrate` **没有 viper 默认值**（缺省即 Go 零值 false），出厂 config_example.yaml
-- 也显式写作 false，因此**按出厂配置部署时本文件的 DDL 同样不会执行**。
--
-- 漏执行 DDL 的后果不是静默降级而是接口 500：生成代码的 SELECT 列清单包含本列
-- （internal/data/ent/project/project.go 的 Columns），缺列会让项目相关查询直接报
-- `Error 1054 Unknown column 'deleted_with_namespace'`。
-- 上线顺序必须 **先执行本迁移、再发新版本**；未开启 db_auto_migrate 的环境由 ops 手工执行。
