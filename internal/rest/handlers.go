package rest

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go_final_project/internal/services"
)

type Task struct {
	Id      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type RespJson struct {
	Id  string `json:"id,omitempty"`
	Err string `json:"error,omitempty"`
}

var task Task

func checkTask(task Task) (RespJson, Task, error) {
	var respJson RespJson
	if task.Title == "" {
		respJson.Err = "Не указан заголовок задачи"
		return respJson, task, fmt.Errorf("Не указан заголовок задачи")
	}
	if task.Date == "" {
		task.Date = time.Now().Format("20060102")
	}
	valTime, err := time.Parse("20060102", task.Date)
	if err != nil {
		respJson.Err = "Дата представлена в формате, отличном от 20060102"
		return respJson, task, fmt.Errorf("Дата представлена в формате, отличном от 20060102")
	}
	if valTime.Day() == time.Now().Day() && valTime.Month() == time.Now().Month() && valTime.Year() == time.Now().Year() {
		task.Date = time.Now().Format("20060102")
		return respJson, task, nil
	}
	if valTime.Before(time.Now()) {
		if task.Repeat == "" {
			task.Date = time.Now().Format("20060102")
		} else {
			task.Date, err = services.NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				respJson.Err = err.Error()
				return respJson, task, fmt.Errorf("Ошибка в ходе вычисления следующей даты")
			}
		}
	}
	return respJson, task, nil
}

func retResponse(res http.ResponseWriter, rsp RespJson) {
	dataJson, err := json.Marshal(rsp)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	res.Header().Set("Content-Type", "application/json; charset=UTF-8")
	res.WriteHeader(http.StatusOK)
	_, _ = res.Write(dataJson)
}

func ApiTaskHandlerPost(conn *sql.DB) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		var respJson RespJson
		var buf bytes.Buffer
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			log.Printf("apiTaskHandler: Ошибка при чтении тела запроса. Error: %s", err)
			respJson.Err = "Ошибка чтения тела запроса"
			retResponse(res, respJson)
			return
		}
		if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
			log.Printf("apiTaskHandler: Ошибка при десериализации JSON. Error: %s", err)
			respJson.Err = "Ошибка десериализации JSON"
			retResponse(res, respJson)
			return
		}
		respJson, task, err = checkTask(task)
		if err != nil {
			retResponse(res, respJson)
			return
		}
		_, err = conn.Exec("INSERT INTO scheduler (date, title,comment,repeat) VALUES (:date, :title, :comment, :repeat);",
			sql.Named("date", task.Date),
			sql.Named("title", task.Title),
			sql.Named("comment", task.Comment),
			sql.Named("repeat", task.Repeat))
		if err != nil {
			log.Printf("apiTaskHandler: Ошибка при добавлении записи в таблицу. Error: %s", err)
			respJson.Err = fmt.Sprintf("Ошибка в ходе вставки данных в БД")
			retResponse(res, respJson)
			return
		}
		var retSelect string
		row := conn.QueryRow("SELECT id FROM scheduler ORDER BY id DESC LIMIT 1;")
		err = row.Scan(&retSelect)
		if err != nil {
			log.Printf("apiTaskHandler: Ошибка при получении ID. Error: %s", err)
			respJson.Err = fmt.Sprintf("Ошибка в ходе получения ID из базы данных")
			retResponse(res, respJson)
			return
		}
		respJson.Id = retSelect
		retResponse(res, respJson)
		return
	}
}

func ApiTaskHandlerGet(conn *sql.DB) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		id := req.URL.Query().Get("id")
		row := conn.QueryRow(fmt.Sprintf("SELECT * FROM scheduler WHERE id = %s", id))
		err := row.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			ret := RespJson{"", "Задача не найдена"}
			dataJson, err := json.Marshal(ret)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
			res.Header().Set("Content-Type", "application/json; charset=UTF-8")
			res.WriteHeader(http.StatusOK)
			_, _ = res.Write(dataJson)
			return
		}
		dataJson, err := json.Marshal(task)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		res.Header().Set("Content-Type", "application/json; charset=UTF-8")
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write(dataJson)
	}
}

func ApiTaskHandlerDelete(conn *sql.DB) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		id := req.URL.Query().Get("id")
		if id == "" {
			ret := RespJson{"", "ID пустой"}
			retResponse(res, ret)
			return
		}
		retRow, err := conn.Exec("DELETE FROM scheduler WHERE id = :id", sql.Named("id", id))
		if err != nil {
			log.Printf("apiTaskHandler: Ошибка в ходе удаления. Error: %s", err)
			ret := RespJson{"", "Ошибка в ходе удаления"}
			retResponse(res, ret)
			return
		}
		rows, err := retRow.RowsAffected()
		if err != nil {
			log.Printf("apiTaskHandler: Ошибка при DELETE в БД. Error: %s", err)
			ret := RespJson{"", "Ошибка в ходе удаления записи в БД"}
			retResponse(res, ret)
			return
		}
		if rows != 1 {
			ret := RespJson{"", "Ошибка в ходе удаления записи в БД"}
			retResponse(res, ret)
			return
		}
		ret := RespJson{"", ""}
		retResponse(res, ret)
		return
	}
}

func ApiTaskHandlerPut(conn *sql.DB) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		var respJson RespJson
		var buf bytes.Buffer
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			log.Printf("apiTaskHandler: Ошибка при чтении тела запроса. Error: %s", err)
			respJson.Err = "Ошибка чтения тела запроса"
			retResponse(res, respJson)
			return
		}
		if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
			log.Printf("apiTaskHandler: Ошибка при десериализации JSON. Error: %s", err)
			ret := RespJson{"", "Ошибка при десериализации JSON"}
			retResponse(res, ret)
			return
		}
		respJson, task, err = checkTask(task)
		if err != nil {
			retResponse(res, respJson)
			return
		}
		retRow, err := conn.Exec("UPDATE scheduler SET date = :date,title = :title,comment=:comment,repeat=:repeat WHERE id = :id",
			sql.Named("date", task.Date),
			sql.Named("title", task.Title),
			sql.Named("comment", task.Comment),
			sql.Named("repeat", task.Repeat),
			sql.Named("id", task.Id))
		rows, err := retRow.RowsAffected()
		if err != nil {
			log.Printf("apiTaskHandler: Ошибка при UPDATE в БД. Error: %s", err)
			ret := RespJson{"", "Ошибка в ходе изменения записи в БД"}
			retResponse(res, ret)
			return
		}
		if rows != 1 {
			ret := RespJson{"", "Ошибка в ходе изменения записи в БД"}
			retResponse(res, ret)
			return
		}
		ret := RespJson{"", ""}
		retResponse(res, ret)
		return
	}
}

func ApiNextDateHandler(conn *sql.DB) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {

		now := req.URL.Query().Get("now")
		date := req.URL.Query().Get("date")
		repeat := req.URL.Query().Get("repeat")
		timeNow, err := time.Parse("20060102", now)
		if err != nil {
			log.Printf("apiNextDateHandler: ошибка парсинга. Парметр запроса now: \"%s\"", now)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		valNextDate, err := services.NextDate(timeNow, date, repeat)
		if err != nil {
			log.Printf("apiNextDateHandler: ошибка функции NextDate. Error: \"%s\"", err)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write([]byte(valNextDate))
		return

	}
}

func ApiTasksHandler(conn *sql.DB) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		rows, err := conn.Query("SELECT * FROM scheduler ORDER BY date LIMIT 50;")
		if err != nil {
			log.Fatal("Ошибка выгрузки данных")
			return
		}
		defer rows.Close()

		var resp []byte

		var tasks []Task
		for rows.Next() {
			task := Task{}

			err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
			if err != nil {
				log.Fatal("Ошибка при выгрузке данных rows.Next()")
				return
			}
			tasks = append(tasks, task)
		}
		if err = rows.Err(); err != nil {
			log.Fatal("Ошибка при выгрузке данных rows.Next()")
			return
		}
		if len(tasks) == 0 {
			m := map[string][]string{"tasks": {}}
			resp, err = json.Marshal(m)
		} else {
			m := make(map[string][]Task)
			m["tasks"] = tasks
			resp, err = json.Marshal(m)
		}

		if err != nil {
			log.Fatal("Ошибка серилизации")
			return
		}
		res.Header().Set("Content-Type", "application/json; charset=UTF-8")
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write(resp)
	}
}

func ApiTaskDone(conn *sql.DB) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		id := req.URL.Query().Get("id")
		if id == "" {
			ret := RespJson{"", "ID пустой"}
			retResponse(res, ret)
			return
		}
		row := conn.QueryRow(fmt.Sprintf("SELECT * FROM scheduler WHERE id = %s", id))
		err := row.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			ret := RespJson{"", "Задача не найдена"}
			retResponse(res, ret)
			return
		}
		if task.Repeat == "" {
			_, err = conn.Exec("DELETE FROM scheduler WHERE id = :id", sql.Named("id", id))
			if err != nil {
				ret := RespJson{"", "Задача не найдена"}
				retResponse(res, ret)
				return
			}
			ret := RespJson{"", ""}
			retResponse(res, ret)
			return
		}
		task.Date, err = services.NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			log.Printf("apiTaskDone: ошибка функции NextDate. TASK: %s", task)
			ret := RespJson{"", "Ошибка при вычислении следующей даты"}
			retResponse(res, ret)
			return
		}
		_, err = conn.Exec("UPDATE scheduler SET date = :date WHERE id = :id",
			sql.Named("date", task.Date),
			sql.Named("id", id))
		if err != nil {
			log.Printf("apiTaskDone: ошибка при запросе БД UPDATE. ID: %s. task.Date: %s", id, task.Date)
			ret := RespJson{"", "Ошибка при вычислении следующей даты"}
			retResponse(res, ret)
			return
		}
		ret := RespJson{"", ""}
		retResponse(res, ret)

	}
}
