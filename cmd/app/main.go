package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"go_final_project/internal/database"
	"go_final_project/internal/rest"
	"go_final_project/tests"

	"github.com/go-chi/chi/v5"
)

func main() {
	database.DefinePath()
	if database.NotExistFile(database.PathDbFile) == true {
		err := database.CreateDbFile()
		if err != nil {
			log.Fatal(err)
			return
		}
	}
	Db, err := sql.Open("sqlite", database.PathDbFile)

	if err != nil {
		log.Fatal(err)
		return
	}
	err = database.CheckTable(Db)
	if err != nil {
		return
	}

	defer Db.Close()
	port, exists := os.LookupEnv("TODO_PORT")
	if !exists {
		port = ":" + strconv.Itoa(tests.Port)
	}
	fmt.Printf("Сервер запущен. Порт: %s", port)

	r := chi.NewRouter()

	r.Get(`/api/nextdate`, rest.ApiNextDateHandler(Db))
	r.Post(`/api/task`, rest.ApiTaskHandlerPost(Db))
	r.Get(`/api/task`, rest.ApiTaskHandlerGet(Db))
	r.Delete(`/api/task`, rest.ApiTaskHandlerDelete(Db))
	r.Put(`/api/task`, rest.ApiTaskHandlerPut(Db))
	r.Get(`/api/tasks`, rest.ApiTasksHandler(Db))
	r.Post(`/api/task/done`, rest.ApiTaskDone(Db))
	r.Handle("/*", http.FileServer(http.Dir("web")))

	if err := http.ListenAndServe(port, r); err != nil {
		fmt.Println("Start server error")
	}
}
