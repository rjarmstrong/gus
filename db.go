package gus

import (
	"database/sql"
	"fmt"
	"strings"
)

const (
	ERR_STRING_NO_ROWS         = "sql: no rows in result set"
	SortDirAsc         SortDir = "ASC"
	SortDirDesc        SortDir = "DESC"
)

type DbOptions struct {
	DriverName     string   // Optional will use sqlite3 by default.
	DataSourceName string   // Optional will use './gus.db' by default.
	Seed           bool     // Caution will regenerate schema and delete data.
	SeedSql        []string // Additional DDL or seed data.
}

func GetDb(o DbOptions) *sql.DB {
	if o.DriverName == "" {
		o.DriverName = "sqlite3"
	}
	if o.DataSourceName == "" {
		o.DataSourceName = "./gus.db"
	}
	db, err := sql.Open(o.DriverName, o.DataSourceName)
	if err != nil {
		panic(err)
	}
	if o.Seed {
		err := Seed(db, o.SeedSql...)
		if err != nil {
			panic(err)
		}
	}
	return db
}

func Seed(db *sql.DB, xtraSeedSql ...string) error {
	_, err := db.Exec(fmt.Sprintf("%s\n%s", DDL, strings.Join(xtraSeedSql, "\n")))
	if err != nil {
		return err
	}
	return nil
}

func CheckNotFound(err error) error {
	if err != nil {
		if err.Error() == ERR_STRING_NO_ROWS {
			return ErrNotFound()
		}
		return err
	}
	return nil
}

func CheckUpdated(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	a, err := res.RowsAffected()
	if a < 1 {
		return ErrNotFound()
	}
	return nil
}

type SortDir string

type ListParams struct {
	Size    int `json:"size"`
	Page    int `json:"page"`
	SortBy  string `json:"sort_by"`
	SortDir SortDir `json:"direction"`
}

func (p *ListParams) ApplyDefaults(){
	if p.Size < 1 {
		p.Size = 20
	}
	if p.SortBy == "" {
		p.SortBy = "updated"
	}
	if string(p.SortDir) == "" {
		p.SortDir = SortDirDesc
	}
}

// GetRows returns a *sql.Rows iterator after adding limit and offset, results are sorted by default 'updated' desc.
// Sql added sample: + ' ORDER by updated DESC LIMIT 20 OFFSET 1'
func GetRows(db *sql.DB, query string, lp ListParams, args ...interface{}) (*sql.Rows, error) {
	lp.ApplyDefaults()
	query += fmt.Sprintf(" ORDER by %s %s", lp.SortBy, lp.SortDir)
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", lp.Size, lp.Page*lp.Size)
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	return rows, err
}

const DDL = `
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email VARCHAR(128) NULL,
    first_name VARCHAR(128) NULL,
    last_name VARCHAR(128) NULL,
    phone VARCHAR(30) NULL,
    password_hash VARCHAR(256) NULL,
    org_id INT,
    updated DATE NOT NULL,
    created DATE NOT NULL,
    deleted BIT,
    role INT,
    CONSTRAINT UC_Email UNIQUE (email)
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

DROP TABLE IF EXISTS orgs;
CREATE TABLE orgs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(128) NOT NULL,
    type INT,
    created DATE NOT NULL,
    updated DATE NOT NULL,
    deleted BIT
);

`
