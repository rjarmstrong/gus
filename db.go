package gus

import (
	"database/sql"
	"fmt"
	"strings"
	"regexp"
)

const (
	ERR_STRING_NO_ROWS                = "sql: no rows in result set"
	ERR_STRING_NO_SUCH_COLUMN         = "sql: no such column"
	DirectionAsc              SortDir = "ASC"
	DirectionDesc             SortDir = "DESC"
)

type DbOptions struct {
	DriverName     string   // Optional will use sqlite3 by default.
	DataSourceName string   // Optional will use './gus.db' by default.
	Seed           bool     // Caution will regenerate schema and delete data.
	SeedSql        []string // Additional DDL or seed data.
}

// Gets the sql database handle for the database specified in the DriverName options parameter.
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

// Seed executes sql prior to app start. Not to be exposed to client apis.
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

type ListArgs struct {
	Size      int `json:"size"`          // Page size
	Page      int `json:"page"`          // Zero-indexed page
	OrderBy   string `json:"sort_by"`    // Order by comma separated columns
	Direction SortDir `json:"direction"` // ASC or DESC
}

func (p *ListArgs) ApplyDefaults() {
	if p.Size < 1 {
		p.Size = 100
	}
	if p.Page < 0 {
		p.Page = 0
	}
	if p.OrderBy == "" {
		p.OrderBy = "updated"
	}
	if string(p.Direction) == "" {
		p.Direction = DirectionDesc
	}
}

var sqlCheck = regexp.MustCompile("^[A-Za-z.]+$")
var sqlErr = ErrInvalid("Invalid order params.")

// GetRows returns a *sql.Rows iterator after adding limit and offset, results are sorted by default 'updated' desc.
// Sql added sample: + ' ORDER by updated DESC LIMIT 20 OFFSET 1'
func GetRows(db *sql.DB, query string, lp ListArgs, args ...interface{}) (*sql.Rows, error) {
	lp.ApplyDefaults()
	if !sqlCheck.MatchString(lp.OrderBy) || !sqlCheck.MatchString(string(lp.Direction)) {
		return nil, sqlErr
	}
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT ? OFFSET ?", lp.OrderBy, lp.Direction)
	args = append(args, lp.Size, lp.Page*lp.Size)
	Debug("LIST:", query, args, lp.Direction)
	stmt, err := db.Prepare(query)
	if err != nil {
		if err.Error() == ERR_STRING_NO_SUCH_COLUMN {
			return nil, ErrInvalid(fmt.Sprintf(err.Error()))
		}
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
    suspended BIT,
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
