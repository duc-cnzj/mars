-- Modify "events" table
ALTER TABLE `events` ADD COLUMN `operator_email` varchar(255) NOT NULL DEFAULT "" AFTER `username`, ADD INDEX `event_operator_email_created_at` (`operator_email`, `created_at`);
