-- Modify "namespaces" table
ALTER TABLE `namespaces` ADD COLUMN `private` bool NOT NULL DEFAULT 0, ADD COLUMN `creator_email` varchar(50) NOT NULL;
-- Create "members" table
CREATE TABLE `members` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime NULL,
  `email` varchar(50) NOT NULL,
  `namespace_id` bigint NULL,
  PRIMARY KEY (`id`),
  INDEX `member_email` (`email`),
  INDEX `members_namespaces_members` (`namespace_id`),
  CONSTRAINT `members_namespaces_members` FOREIGN KEY (`namespace_id`) REFERENCES `namespaces` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL
) CHARSET utf8mb4 COLLATE utf8mb4_bin;
