# go-ddls2m
Lazy DDL converter from Cloud Spanner to MySQL for document.
This is not for production use, just for document.(ie ER diagram input)

# how to run
go run cmd/ddls2m/main.go -f spanner_ddl.sql

# convert details

- convert only create table, create index DDL
  - not support other DDL 
- convert column type https://cloud.google.com/solutions/migrating-mysql-to-spanner?hl=ja#supported_data_types
- convert interleave to foreigin key
- table engine is InnoDB, charset is utf8mb4.

# example
```sql
cat spanner_ddl.sql
CREATE TABLE `users` (
	`user_id` STRING(36) NOT NULL,
	`name` STRING(MAX) NOT NULL,
	`uid` STRING(255) NOT NULL,
	`created_at` TIMESTAMP NOT NULL,
	`updated_at` TIMESTAMP NOT NULL
) PRIMARY KEY  (`user_id`);
CREATE UNIQUE INDEX idx_users_uid ON users (uid);
CREATE TABLE `friends` (
	`friend_id` INT64 NOT NULL,
	`user_id` STRING(36) NOT NULL,
	`to_id` STRING(36) NOT NULL,
	`created_at` TIMESTAMP NOT NULL,
	`updated_at` TIMESTAMP NOT NULL
) PRIMARY KEY  (`user_id`, `friend_id`),
INTERLEAVE IN PARENT `users` ;
CREATE UNIQUE INDEX idx_friends_user_id_to_id ON friends (user_id, to_id);
CREATE  INDEX idx_friends_to_id ON friends (to_id);
```
```
go run cmd/ddls2m/main.go -s spanner_ddl.sql
CREATE TABLE `users` (
	`user_id` VARCHAR(36) NOT NULL,
	`name` TEXT NOT NULL,
	`uid` VARCHAR(255) NOT NULL,
	`created_at` DATETIME NOT NULL,
	`updated_at` DATETIME NOT NULL,
	PRIMARY KEY  (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE UNIQUE INDEX idx_users_uid ON users(uid);
CREATE TABLE `friends` (
	`friend_id` BIGINT NOT NULL,
	`user_id` VARCHAR(36) NOT NULL,
	`to_id` VARCHAR(36) NOT NULL,
	`created_at` DATETIME NOT NULL,
	`updated_at` DATETIME NOT NULL,
	PRIMARY KEY  (`user_id`, `friend_id`),
	FOREIGN KEY  (`user_id`) REFERENCES users (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE UNIQUE INDEX idx_friends_user_id_to_id ON friends(user_id, to_id);
CREATE INDEX idx_friends_to_id ON friends(to_id);
```
