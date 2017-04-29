package gus

import (
	//_ "github.com/mattn/go-sqlite3"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"runtime/debug"
	"testing"
	"time"
	"fmt"
)

var orgsv *Orgs
var us *Users

func TestMain(m *testing.M) {
	// docker run --rm --name gus-db -e MYSQL_ALLOW_EMPTY_PASSWORD=yes -p 127.0.0.1:3306:3306 -d mysql:5.7.18
	conn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:%s)/gus?parseTime=true&multiStatements=true", "root", "", "3306")
	db, err := GetDb(DbOpts{Seed: true, DataSourceName:conn, DriverName:"mysql"})
	defer db.Close()
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
