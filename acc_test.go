package gus

import (
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"os"
	"testing"
	"runtime/debug"
)

var ps *Orgs
var us *Users
var db *sql.DB

func TestMain(m *testing.M) {
	db = GetDb(DbOptions{Seed: true})
	ps = NewOrgs(db)
	us = NewUsers(db)
	code := m.Run()
	os.Exit(code)
}

func ErrIf(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err, string(debug.Stack()))
	}
}
