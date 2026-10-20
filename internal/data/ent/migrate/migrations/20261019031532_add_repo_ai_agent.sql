-- Modify "repos" table
ALTER TABLE `repos` ADD COLUMN `ai_agent` bool NOT NULL DEFAULT 0 AFTER `description`, ADD COLUMN `ai_agent_key` varchar(255) NULL, ADD UNIQUE INDEX `ai_agent_key` (`ai_agent_key`);
