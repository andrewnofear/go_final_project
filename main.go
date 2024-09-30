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

var Db *sql.DB

func main() {
	definePath()
	if notExistFile(PathDbFile) == true { //ЗДЕСЬ, если файл не существует - вставка выполнится
		err := createDbFile()
		if err != nil {
			log.Fatal(err)
			return
		}
	}
	Db, err := sql.Open("sqlite", PathDbFile)
	defer Db.Close()

	if err != nil {
		log.Fatal(err)
		return
	}

	if rows, tableCheck := Db.Query("SELECT * FROM scheduler;"); tableCheck != nil { //проверка, если нет таблиц - создаем
		err = CreateTable(Db)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer rows.Close()
	}

	//ДЛЯ ТЕСТА. Первый запуск - все ок, повторный, SQL busy
	rows, err := Db.Query("INSERT INTO scheduler (date, title,comment,repeat) VALUES ('1', '2', '3', '4');")
	defer rows.Close()
	//ТЕСТ ЗАВЕРШЕН

	port, exists := os.LookupEnv("TODO_PORT")
	if !exists {
		port = ":" + strconv.Itoa(tests.Port)
	}
	fmt.Println(port)
	http.Handle("/", http.FileServer(http.Dir("web")))
	http.HandleFunc(`/api/nextdate`, apiNextDateHandler)
	http.HandleFunc(`/api/task`, apiTaskHandler)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Println("Start server error")
	}
}
