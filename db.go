package gus

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	ErrStringNoSuchColumn         = "sql: no such column"
	DirectionAsc          SortDir = "ASC"
	DirectionDesc         SortDir = "DESC"
)

func Milliseconds(t time.Time) int64 {
	u := int64(time.Duration(t.UnixNano()) / time.Millisecond)
	return u
}

type DbOpts struct {
	DriverName     string   // Optional will use sqlite3 by default.
	DataSourceName string   // Optional will use './gus.db' by default.
	Seed           bool     // Caution will regenerate schema and delete data.
	SeedSql        []string // Additional DDL or seed data.
}

var (
	driverName string
	sqlCheck   = regexp.MustCompile("^[A-Za-z.]+$")
	sqlErr     = ErrInvalid("Invalid order params.")
	seeds      = map[string]string{
		"mysql":   SeedMySql,
		"sqlite3": SeedSqlLite,
	}
)

// Gets the sql database handle for the database specified in the DriverName options parameter.
func GetDb(o DbOpts) (*sql.DB, error) {
	if o.DriverName == "" {
		o.DriverName = "sqlite3"
	}
	driverName = o.DriverName
	db, err := sql.Open(o.DriverName, o.DataSourceName)
	if err != nil {
		return nil, err
	}
	if o.Seed {
		err := Seed(db, o.SeedSql...)
		if err != nil {
			return nil, err
		}
	}
	return db, nil
}

// Seed executes sql prior to app start. Not to be exposed to client apis.
func Seed(db *sql.DB, xtraSeedSql ...string) error {
	_, err := db.Exec(fmt.Sprintf("%s\n%s", seeds[driverName], strings.Join(xtraSeedSql, "\n")))
	if err != nil {
		return err
	}
	return nil
}

func CheckNotFound(err error) error {
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
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
		return ErrNotFound
	}
	return nil
}

type SortDir string

type ListArgs struct {
	Size      int     `json:"size"`      // Page size
	Page      int     `json:"page"`      // Zero-indexed page
	OrderBy   string  `json:"sort_by"`   // Order by comma separated columns
	Direction SortDir `json:"direction"` // ASC or DESC
	Deleted   bool    `json:"deleted"`   // Include deleted in results
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

// ApplyUpdates will apply updates to an 'original' struct and update fields based on an 'updates' struct
// The 'updates' struct should have point fields and should also serialize to and from json the same as the
// Intended destination fields.
func ApplyUpdates(original interface{}, updates interface{}) error {
	p, err := json.Marshal(updates)
	if err != nil {
		return err
	}
	json.Unmarshal(p, original)
	return nil
}

// GetRows returns a *sql.Rows iterator after adding limit and offset, results are sorted by default 'updated' desc.
// Sql added sample: + ' ORDER by updated DESC LIMIT 20 OFFSET 1'
func GetRows(db *sql.DB, query string, lp *ListArgs, args ...interface{}) (*sql.Rows, error) {
	lp.ApplyDefaults()
	if !sqlCheck.MatchString(lp.OrderBy) || !sqlCheck.MatchString(string(lp.Direction)) {
		return nil, sqlErr
	}
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT ? OFFSET ?", lp.OrderBy, lp.Direction)
	args = append(args, lp.Size, lp.Page*lp.Size)
	stmt, err := db.Prepare(query)
	if err != nil {
		if err.Error() == ErrStringNoSuchColumn {
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

func Tx(db *sql.DB, txFunc func(*sql.Tx) error) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	err = txFunc(tx)
	return err
}
