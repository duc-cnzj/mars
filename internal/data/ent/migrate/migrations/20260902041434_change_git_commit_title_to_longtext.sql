-- Modify "changelogs" table
ALTER TABLE `changelogs` MODIFY COLUMN `git_commit_title` longtext NULL;
-- Modify "projects" table
ALTER TABLE `projects` MODIFY COLUMN `git_commit_title` longtext NULL;
