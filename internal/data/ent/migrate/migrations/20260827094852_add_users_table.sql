-- Create "users" table
CREATE TABLE `users` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `email` varchar(255) NOT NULL,
  `name` varchar(255) NOT NULL DEFAULT "",
  `roles` json NOT NULL,
  `last_login` timestamp NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `email` (`email`),
  INDEX `user_last_login` (`last_login`)
) CHARSET utf8mb4 COLLATE utf8mb4_bin;
