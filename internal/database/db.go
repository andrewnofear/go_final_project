package database

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

type Task struct {
	Id      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func NotExistFile(fileName string) bool {
	_, err := os.Stat(PathDbFile)
	if err != nil {
		return true
	}
	return false
}

func DefinePath() error {
	AppPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return err
	}
	PathDbFile = filepath.Join(AppPath, "scheduler.db")
	return nil
}

func CheckTable(conn *sql.DB) error {
	row, tableCheck := conn.Query("SELECT * FROM scheduler LIMIT 1;")
	if tableCheck != nil {
		err := CreateTable(conn)
		if err != nil {
			log.Fatal(err)
			return err
		}
	} else {
		row.Close()
	}
	return nil
}

func CreateDbFile() error {
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
