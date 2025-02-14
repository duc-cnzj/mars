-- Modify "access_tokens" table
ALTER TABLE `access_tokens` MODIFY COLUMN `email` varchar(50) NOT NULL COLLATE utf8mb4_general_ci;
-- Modify "favorites" table
ALTER TABLE `favorites` MODIFY COLUMN `email` varchar(50) NOT NULL COLLATE utf8mb4_general_ci;
-- Modify "members" table
ALTER TABLE `members` MODIFY COLUMN `email` varchar(50) NOT NULL COLLATE utf8mb4_general_ci;
-- Modify "namespaces" table
ALTER TABLE `namespaces` MODIFY COLUMN `creator_email` varchar(50) NOT NULL COLLATE utf8mb4_general_ci;
