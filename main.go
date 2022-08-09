package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

/*
GET    /tasks      => Get all tasks
GET    /tasks/{id} => Get one task
POST   /tasks      => Create new task
PUT    /tasks/{id} => Updates all values of a task
DELETE /tasks/{id} => Deletes a given task
*/

var db *sql.DB

type Response struct {
	Error   bool        `json:"error"`
	Message interface{} `json:"message,omitempty"`
}

func newResponse(error bool, message interface{}) []byte {
	resp := Response{
		Error:   error,
		Message: message,
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		fmt.Println(err.Error())
	}
	return respBytes
}

func JSONError(w http.ResponseWriter, err string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, err)
}

type Task struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt int    `json:"createdAt"`
	IsDone    bool   `json:"isDone"`
}

var regexps = map[string]*regexp.Regexp{
	"tasksRe": regexp.MustCompile(`^/tasks/?$`),
	"taskRe":  regexp.MustCompile(`^/tasks/(\d+)$`),
}

func tasksHandler(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	switch {
	case req.Method == http.MethodOptions:
		return
	case regexps["tasksRe"].MatchString(path) && req.Method == http.MethodGet:
		getTasks(w)
	case regexps["tasksRe"].MatchString(path) && req.Method == http.MethodPost:
		createTask(w, req)
	case regexps["taskRe"].MatchString(path) && req.Method == http.MethodPut:
		updateTask(w, req)
	case regexps["taskRe"].MatchString(path) && req.Method == http.MethodDelete:
		deleteTask(w, req)
	}
}

func deleteTask(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	matches := regexps["taskRe"].FindStringSubmatch(req.URL.Path)
	if len(matches) < 2 {
		w.WriteHeader(http.StatusNotFound)
		w.Write(newResponse(false, ""))
		JSONError(w, string(newResponse(true, "No such task")), http.StatusNotFound)
		return
	}
	taskId, err := strconv.Atoi(matches[1])
	if err != nil {
		JSONError(w, string(newResponse(true, err.Error())), http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}

	var task Task
	err = db.QueryRow("SELECT id, name FROM todos WHERE id=$1", taskId).
		Scan(&task.Id, &task.Name)
	if err != nil {
		fmt.Println(err.Error())
		JSONError(w, string(newResponse(true, "No such task")), http.StatusNotFound)
		return
	}

	stmt, err := db.Prepare("DELETE FROM todos WHERE id=$1")
	if err != nil {
		JSONError(w, string(newResponse(true, err.Error())), http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(task.Id)
	if err != nil {
		JSONError(w, string(newResponse(true, err.Error())), http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}
	w.Write(newResponse(false, nil))
}

func updateTask(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	matches := regexps["taskRe"].FindStringSubmatch(req.URL.Path)
	if len(matches) < 2 {
		JSONError(w, string(newResponse(true, "No such task")), http.StatusNotFound)
		return
	}
	taskId, err := strconv.Atoi(matches[1])
	if err != nil {
		JSONError(w, string(newResponse(true, err.Error())), http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}
	var task Task
	err = db.QueryRow("SELECT * FROM todos WHERE id=$1", taskId).
		Scan(&task.Id, &task.Name, &task.IsDone, &task.CreatedAt)
	if err != nil {
		JSONError(w, string(newResponse(true, "No such task")), http.StatusNotFound)
		return
	}

	dec := json.NewDecoder(req.Body)
	var t Task
	err = dec.Decode(&t)
	if err != nil {
		JSONError(w, string(newResponse(true, "Bad JSON")), http.StatusBadRequest)
		fmt.Println(err.Error())
		return
	}
	stmt, err := db.Prepare("UPDATE todos SET name=$1, is_done=$2 WHERE id=$3")
	if err != nil {
		JSONError(w, string(newResponse(true, err.Error())), http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(t.Name, t.IsDone, task.Id)
	if err != nil {
		JSONError(w, string(newResponse(true, err.Error())), http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}
	w.Write(newResponse(false, nil))
}

func createTask(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dec := json.NewDecoder(req.Body)
	var t Task
	err := dec.Decode(&t)
	if err != nil {
		JSONError(w, string(newResponse(true, "Bad JSON")), http.StatusBadRequest)
		fmt.Println(err.Error())
		return
	}
	stmt, err := db.Prepare("INSERT INTO todos(name, created_at, is_done) VALUES($1, $2, $3)")
	if err != nil {
		JSONError(w, string(newResponse(true, err.Error())), http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(t.Name, time.Now().UTC().Unix(), false)
	if err != nil {
		JSONError(w, string(newResponse(true, err.Error())), http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}
	w.Write(newResponse(false, nil))
}

func getTasks(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	rows, err := db.Query("SELECT * FROM todos")
	if err != nil {
		JSONError(w, string(newResponse(true, err.Error())), http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}
	var tasks []Task
	for rows.Next() {
		var task Task
		err = rows.Scan(&task.Id, &task.Name, &task.IsDone, &task.CreatedAt)
		if err != nil {
			JSONError(w, string(newResponse(true, err.Error())), http.StatusInternalServerError)
			fmt.Println(err.Error())
			return
		}
		tasks = append(tasks, task)
	}
	err = rows.Err()
	if err != nil {
		JSONError(w, string(newResponse(true, err.Error())), http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}
	w.Write(newResponse(false, tasks))
}

func indexHandler(w http.ResponseWriter, req *http.Request) {

	if req.URL.Path == "/favicon.ico" {
		http.ServeFile(w, req, "./client/dist/favicon.ico")
		return
	}
	http.ServeFile(w, req, "./client/dist/index.html")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		next.ServeHTTP(w, req)
	})
}

func main() {
	var err error
	err = godotenv.Load(".env")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DBUSER"), os.Getenv("DBPASSWORD"), os.Getenv("DBNAME"))
	db, err = sql.Open("postgres", dbinfo)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.Handle("/tasks/", corsMiddleware(http.HandlerFunc(tasksHandler)))
	// статика
	mux.HandleFunc("/", indexHandler)
	mux.Handle("/static/", http.FileServer(http.Dir("./client/dist")))
	mux.Handle("/test/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("test"))
	}))

	err = http.ListenAndServe(":8081", mux)
	if err != nil {
		fmt.Println(err.Error())
	}
}
