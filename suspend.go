package gus

import (
	"time"
	"database/sql"
	"fmt"
)

func NewSuspender(table string, db *sql.DB) *Suspender {
	return &Suspender{table: table, db: db}
}

type Suspender struct {
	table string
	db    *sql.DB
}

func (su *Suspender) Suspend(id int64) error {
	stmt, err := su.db.Prepare(fmt.Sprintf("UPDATE %s SET suspended = 1, updated = ? WHERE id = ? AND deleted = 0", su.table))
	if err != nil {
		return err
	}
	return CheckUpdated(stmt.Exec(time.Now(), id))
}

func (su *Suspender) Restore(id int64) error {
	stmt, err := su.db.Prepare(fmt.Sprintf("UPDATE %s SET suspended = 0, updated = ? WHERE id = ? AND deleted = 0", su.table))
	if err != nil {
		return err
	}
	return CheckUpdated(stmt.Exec(time.Now(), id))
}
