package gus

const SeedMySql = `

DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    uid VARCHAR(36) NULL,
    username VARCHAR(128) NULL,
    email VARCHAR(128) NULL,
    first_name VARCHAR(128) NULL,
    last_name VARCHAR(128) NULL,
    phone VARCHAR(30) NULL,
    password_hash VARCHAR(256) NULL,
    invite_code VARCHAR(30) NULL,
    org_id BIGINT,
    updated BIGINT NULL DEFAULT 0,
    created BIGINT NULL DEFAULT 0,
    suspended tinyint(4),
    deleted tinyint(4),
    role BIGINT,
	passive TINYINT(2) NULL,
    CONSTRAINT UC_Email UNIQUE (email),
    CONSTRAINT UC_Username UNIQUE (username)
);

DROP TABLE IF EXISTS password_resets;
CREATE TABLE password_resets (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    email VARCHAR(128) NULL,
    reset_token VARCHAR(256) NULL,
    created BIGINT NULL DEFAULT 0,
    deleted tinyint(4)
);

DROP TABLE IF EXISTS password_attempts;
CREATE TABLE password_attempts (
    username VARCHAR(250),
    created BIGINT NULL DEFAULT 0
);

DROP TABLE IF EXISTS orgs;
CREATE TABLE orgs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(128) NOT NULL,
    street VARCHAR(512) NULL,
    suburb VARCHAR(512) NULL,
    town VARCHAR(512) NULL,
    postcode VARCHAR(512) NULL,
    country VARCHAR(512) NULL,
    type INT,
    created BIGINT NULL DEFAULT 0,
    updated BIGINT NULL DEFAULT 0,
    suspended tinyint(4),
    deleted tinyint(4)
);

`
