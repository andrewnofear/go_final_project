package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Task struct {
	Date    string `json:"date, omitempty"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

type respJson struct {
	id  int64 `json:"id,omitempty"`
	err error `json:"error,omitempty"`
}

var respJ *respJson

func retResp(res http.ResponseWriter) {
	dataJson, err := json.Marshal(respJ)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(dataJson)
}

func apiTaskHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		defer retResp(res)

		var task Task
		var buf bytes.Buffer

		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Println(buf.String())
		if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
			respJ.err = fmt.Errorf("ошибка десериализации JSON")
			fmt.Println("ошибка десериализации JSON")
			return
		}
		if task.Title == "" {
			respJ.err = fmt.Errorf("не указан заголовок задачи")
			fmt.Println("не указан заголовок задачи")
			return
		}
		if task.Date == "" {
			time.Now().Format("20060102")
		} else {
			valTime, err := time.Parse("20060102", task.Date)
			if err != nil {
				respJ.err = fmt.Errorf("дата представлена в формате, отличном от 20060102")
				fmt.Println("дата представлена в формате, отличном от 20060102")
				return
			}
			if valTime.Before(time.Now()) {
				if task.Repeat == "" {
					task.Date = time.Now().Format("20060102")
				} else {
					task.Date, err = NextDate(time.Now(), task.Date, task.Repeat)
					if err != nil {
						respJ.err = fmt.Errorf("правило повторения указано в неправильном формате")
						fmt.Println("правило повторения указано в неправильном формате")
						return
					}
				}
			}
		}
		respDb, err := Db.Exec("INSERT INTO scheduler (date, title,comment,repeat) VALUES (:date, :title, :comment, :repeat)",
			sql.Named("date", task.Date),
			sql.Named("title", task.Title),
			sql.Named("comment", task.Comment),
			sql.Named("repeat", task.Repeat)) //Вставка не выполняется как и в main

		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		respJ.id, err = respDb.LastInsertId()
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

	} else {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
}

func apiNextDateHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		now := req.URL.Query().Get("now")
		date := req.URL.Query().Get("date")
		repeat := req.URL.Query().Get("repeat")
		timeNow, err := time.Parse("20060102", now)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		} else {
			valNextDate, err := NextDate(timeNow, date, repeat)
			if err != nil {
				res.WriteHeader(http.StatusBadRequest)
				return
			}
			res.WriteHeader(http.StatusOK)
			res.Write([]byte(valNextDate))
		}
	} else {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
}
