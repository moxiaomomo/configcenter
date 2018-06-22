package mysql

var (
	configsetSchema = `
CREATE TABLE IF NOT EXISTS configs (
id INT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
name varchar(32) not null,
path varchar(64) not null,
version varchar(16) not null default '1.0',
comment varchar(512) not null default '',
createdAt int(11) default 0,
updatedAt int(11) default 0,
status int(11) default 1,
changeset_timestamp int(11) default 0,
changeset_checksum varchar(255) not null default '',
changeset_data BLOB,
changeset_source varchar(255) not null default '',
changeset_format varchar(32) not null default '',
unique key(name, path, version),
key(createdAt),
key(updatedAt));
`

	configsetLogSchema = `
CREATE TABLE IF NOT EXISTS configs_audit (
id INT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
action varchar(16) not null,
name varchar(32) not null,
path varchar(64) not null,
version varchar(16) not null default '1.0',
comment varchar(512) default '',
createdAt int(11) default 0,
updatedAt int(11) default 0,
status int(11) default 1,
changeset_timestamp int(11) default 0,
changeset_checksum varchar(255) default '',
changeset_data BLOB,
changeset_source varchar(255) default '',
changeset_format varchar(32) default '',
logts TIMESTAMP DEFAULT 0 ON UPDATE CURRENT_TIMESTAMP,
key(name, path, version),
key(createdAt),
key(updatedAt));
`
)
