-- Modify "projects" table
ALTER TABLE `projects` ADD COLUMN `updated_by` varchar(255) NULL AFTER `creator`;
