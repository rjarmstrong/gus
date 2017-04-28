package gus

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"runtime/debug"
	"testing"
	"time"
)

var orgsv *Orgs
var us *Users
var db *sql.DB

func TestMain(m *testing.M) {
	db, err := GetDb(DbOpts{Seed: true})
	if err != nil {
		panic(err)
	}
	orgsv = NewOrgs(db)
	us = NewUsers(db, UserOpts{MaxAuthAttempts: 5, AttemptLockDuration: time.Duration(1) * time.Second})
	code := m.Run()
	os.Exit(code)
}

func ErrIf(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err, string(debug.Stack()))
	}
}
