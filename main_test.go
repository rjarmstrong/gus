package gus

import (
	_ "github.com/go-sql-driver/mysql"
	"os"
	"testing"
	"fmt"
)

var orgsv *Orgs
var us *Users

func TestMain(m *testing.M) {
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:%s)/kwk_test?parseTime=true&multiStatements=true", "root", "rootPassword", "3306")
	db, err := GetDb(DbOpts{Seed: true, DriverName: "mysql", DataSourceName: dsn})
	defer db.Close()
	if err != nil {
		panic(err)
	}
	orgsv = NewOrgs(db)
	us = NewUsers(db, UserOpts{AuthAttempts: 5, AuthLockDuration: 1, ResetTokenExpiry: 1 })
	code := m.Run()
	os.Exit(code)
}
