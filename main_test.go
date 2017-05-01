package gus

import (
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
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:%s)/gus?parseTime=true&multiStatements=true", "root", "rootPassword", "3306")
	db, err := GetDb(DbOpts{Seed: true, DriverName: "mysql", DataSourceName: dsn})
	defer db.Close()
	if err != nil {
		panic(err)
	}
	orgsv = NewOrgs(db)
	us = NewUsers(db, UserOpts{AuthAttempts: 5, AuthLockDuration: time.Duration(1) * time.Second})
	code := m.Run()
	os.Exit(code)
}

func ErrIf(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err, string(debug.Stack()))
	}
}
