package main

import (
	"database/sql"
	"fmt"
	"go_final_project/tests"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	var Db *sql.DB
	definePath()
	if notExistFile(PathDbFile) == true {
		err := createDbFile()
		if err != nil {
			log.Fatal(err)
			return
		}
	}
	Db, err := sql.Open("sqlite", PathDbFile)

	if err != nil {
		log.Fatal(err)
		return
	}
	row, tableCheck := Db.Query("SELECT * FROM scheduler;")
	if tableCheck != nil {
		err = CreateTable(Db)
		if err != nil {
			log.Fatal(err)
			return
		}
	} else {
		row.Close()
	}

	Db.Close()
	port, exists := os.LookupEnv("TODO_PORT")
	if !exists {
		port = ":" + strconv.Itoa(tests.Port)
	}
	fmt.Printf("Сервер запущен. Порт: %s", port)

	http.Handle("/", http.FileServer(http.Dir("web")))
	http.HandleFunc(`/api/nextdate`, apiNextDateHandler)
	http.HandleFunc(`/api/task`, apiTaskHandler)
	http.HandleFunc(`/api/tasks`, apiTasksHandler)
	http.HandleFunc(`/api/task/done`, apiTaskDone)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Println("Start server error")
	}
}
