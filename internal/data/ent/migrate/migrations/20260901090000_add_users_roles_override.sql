-- Modify "users" table
ALTER TABLE `users` ADD COLUMN `roles_override` boolean NOT NULL DEFAULT false;
