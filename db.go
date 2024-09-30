package main

import (
	"database/sql"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

var PathDbFile string
var AppPath string

func notExistFile(fileName string) bool {
	_, err := os.Stat(PathDbFile)

	if err != nil {
		return true
	}
	return false
}

func definePath() error {
	AppPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return err
	}
	PathDbFile = filepath.Join(AppPath, "scheduler.db")
	return nil
}

func createDbFile() error {
	_, err := os.Create(PathDbFile)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func CreateTable(db *sql.DB) error {
	dbFile := filepath.Join(AppPath, "SQLscheduler.sql")
	_, err := os.Stat(dbFile)
	if err != nil {
		log.Fatal(err)
		return err
	}

	file, err := ioutil.ReadFile(dbFile)
	if err != nil {
		log.Fatal(err)
		return err
	}

	commandSql := strings.Split(string(file), ";")
	for _, com := range commandSql {
		_, err = db.Query(com + ";")
	}
	return nil
}
