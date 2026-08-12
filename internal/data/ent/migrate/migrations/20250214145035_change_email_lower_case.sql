UPDATE `access_tokens` set `email` = LOWER(`email`);
UPDATE `favorites` set `email` = LOWER(`email`);
UPDATE `members` set `email` = LOWER(`email`);
UPDATE `namespaces` set `creator_email` = LOWER(`creator_email`);
