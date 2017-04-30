package gus

const SeedSqlLite = `
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid VARCHAR(36) NULL,
    username VARCHAR(128) NULL,
    email VARCHAR(128) NULL,
    first_name VARCHAR(128) NULL,
    last_name VARCHAR(128) NULL,
    phone VARCHAR(30) NULL,
    invite_code VARCHAR(30) NULL,
    password_hash VARCHAR(256) NULL,
    org_id INT,
    updated DATE NOT NULL,
    created DATE NOT NULL,
    suspended BIT,
    deleted BIT,
    role INT,
    CONSTRAINT UC_Email UNIQUE (email),
    CONSTRAINT UC_Username UNIQUE (username)
);

DROP TABLE IF EXISTS password_resets;
CREATE TABLE password_resets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INT NOT NULL,
    email VARCHAR(128) NULL,
    reset_token VARCHAR(256) NULL,
    created DATE NOT NULL,
    deleted BIT
);

DROP TABLE IF EXISTS password_attempts;
CREATE TABLE password_attempts (
    username VARCHAR(250),
    created INT NOT NULL
);

DROP TABLE IF EXISTS orgs;
CREATE TABLE orgs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(128) NOT NULL,
    type INT,
    created DATE NOT NULL,
    updated DATE NOT NULL,
    suspended BIT,
    deleted BIT
);

`
