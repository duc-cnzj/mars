-- Modify "changelogs" table
ALTER TABLE `changelogs` ADD INDEX `changelog_deleted_at_created_at` (`deleted_at`, `created_at`);
