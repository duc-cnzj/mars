-- Modify "project_hibernates" table
ALTER TABLE `project_hibernates` ADD COLUMN `window_names` json NULL AFTER `snapshot`;
