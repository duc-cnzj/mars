-- Modify "favorites" table
ALTER TABLE `favorites` ADD COLUMN `sort_order` bigint NOT NULL DEFAULT 0 AFTER `email`, ADD UNIQUE INDEX `favorite_email_namespace_id` (`email`, `namespace_id`), ADD INDEX `favorite_email_sort_order` (`email`, `sort_order`);
