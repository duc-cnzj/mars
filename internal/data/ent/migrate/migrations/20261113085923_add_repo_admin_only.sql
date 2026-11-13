-- Modify "repos" table
ALTER TABLE `repos` ADD COLUMN `admin_only` bool NOT NULL DEFAULT 0 AFTER `enabled`;
