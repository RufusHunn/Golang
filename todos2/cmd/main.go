package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"text/template"
	"todos2/store"

	"github.com/google/uuid"
)

type contextKey int

const (
	traceCtxKey contextKey = iota + 1
)

var TraceIDString = uuid.New().String()

var Tasks map[string]store.Task = make(map[string]store.Task)

var TaskChan chan TaskMessage = make(chan TaskMessage)

type TaskMessage struct {
	Action   string
	Key      string
	Payload  store.Task
	Response chan interface{}
}

func generateID() string {
	if len(Tasks) == 0 {
		return "1"
	}

	keys := make([]int, 0, len(Tasks))
	for k := range Tasks {
		ik, _ := strconv.Atoi(k)
		keys = append(keys, ik)
	}
	nextKey := keys[len(keys)-1] + 1
	return fmt.Sprintf("%d", nextKey)
}

// Handler layer below

func getItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := getData(r.PathValue("ix"))

		if result == nil {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(result)
		slog.InfoContext(r.Context(), fmt.Sprintf("Retrieved item: %s", r.PathValue("ix")))
	}
}

func addItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var description = r.PathValue("description")
		var status = r.PathValue("status")
		if description != "" && status != "" {
			addData(description, status)
			slog.InfoContext(r.Context(), fmt.Sprintf("Added new item with description: %s", description))
		}
	}
}

func updateItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := updateData(r.PathValue("ix"), r.PathValue("description"), r.PathValue("status"))

		if result == nil {
			http.NotFound(w, r)
			return
		}

		w.WriteHeader(http.StatusOK)
		slog.InfoContext(r.Context(), fmt.Sprintf("Updated item: %s", r.PathValue("ix")))
	}
}

func deleteItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := deleteData(r.PathValue("ix"))

		if result == nil {
			http.NotFound(w, r)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		slog.InfoContext(r.Context(), fmt.Sprintf("Deleted item: %s", r.PathValue("ix")))
	}
}

// Complex Layer Below

func getData(key string) interface{} {
	resp := make(chan interface{})
	TaskChan <- TaskMessage{Action: "read", Key: key, Response: resp}
	result := <-resp
	return result
}

func addData(description, status string) {
	var task = store.Task{Description: description, Status: status}
	id := generateID()
	resp := make(chan interface{})
	TaskChan <- TaskMessage{Action: "create", Key: id, Payload: task, Response: resp}
	<-resp
}

func updateData(id, description, status string) interface{} {
	var task = store.Task{Description: description, Status: status}
	resp := make(chan interface{})
	TaskChan <- TaskMessage{Action: "update", Key: id, Payload: task, Response: resp}
	result := <-resp
	return result
}

func deleteData(key string) interface{} {
	resp := make(chan interface{})
	TaskChan <- TaskMessage{Action: "delete", Key: key, Response: resp}
	result := <-resp
	return result
}

// Data layer below

func create(key string, payload store.Task) {
	Tasks[key] = payload
}

func read(key string) (store.Task, bool) {
	task, ok := Tasks[key]
	return task, ok
}

func del(key string) {
	delete(Tasks, key)
}

func renderHTMLTable(w http.ResponseWriter) {
	const tpl = `
    <!DOCTYPE html>
    <html>
    <head><title>Tasks Table</title></head>
    <body>
        <h1>Tasks</h1>
        <table border="1" cellpadding="6" cellspacing="0">
            <tr><th>ID</th><th>Description</th><th>Status</th></tr>
            {{range $id, $task := .}}
            <tr>
                <td>{{$id}}</td>
                <td>{{$task.Description}}</td>
                <td>{{$task.Status}}</td>
            </tr>
            {{end}}
        </table>
    </body>
    </html>
    `

	tmpl := template.Must(template.New("table").Parse(tpl))
	tmpl.Execute(w, Tasks)
}

func main() {

	ctx := context.Background()
	ctx = context.WithValue(ctx, traceCtxKey, TraceIDString)

	var handler slog.Handler
	handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
	})

	handler = MyHandler{handler}

	slog.SetDefault(slog.New(handler))

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: contextMiddleware(ctx, traceIDMiddleware(mux)),
	}

	if err := srv.Shutdown(ctx); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}

	// Set up TaskMessage channel, load data and run actor goroutine
	Tasks = store.Load()
	go taskActor()

	mux.HandleFunc("/get/{ix}", getItem())
	mux.HandleFunc("/create/{description}/{status}", addItem())
	mux.HandleFunc("/update/{ix}/{description}/{status}", updateItem())
	mux.HandleFunc("/delete/{ix}", deleteItem())

	// Below shows a navigable directory with one static page; is this sufficient?
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/about/", http.StripPrefix("/about", fs))

	// Dynamic webpage
	mux.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		renderHTMLTable(w)
	})

	go func() {
		if err := http.ListenAndServe(":8080", srv.Handler); err != http.ErrServerClosed {
			panic("ListenAndServe: " + err.Error())
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	fmt.Println("App is running. Press Ctrl+C to exit...")
	<-sig
	slog.InfoContext(ctx, "Saving data")
	store.Save(Tasks)
	fmt.Println("Exiting")

}

func taskActor() {

	for msg := range TaskChan {
		switch msg.Action {
		case "create":
			create(msg.Key, msg.Payload)
			msg.Response <- true
		case "read":
			task, ok := read(msg.Key)
			if ok {
				msg.Response <- task
			} else {
				msg.Response <- nil
			}
		case "update":
			_, exists := Tasks[msg.Key]
			if exists {
				create(msg.Key, msg.Payload)
				msg.Response <- true
			} else {
				msg.Response <- nil
			}
		case "delete":
			_, exists := Tasks[msg.Key]
			if exists {
				del(msg.Key)
				msg.Response <- true
			} else {
				msg.Response <- nil
			}
		case "list":
			msg.Response <- Tasks
		}
	}
}

type MyHandler struct {
	slog.Handler
}

func (h MyHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceID, ok := ctx.Value(traceCtxKey).(string); ok {
		r.Add("trace_id", slog.StringValue(traceID))
	}

	return h.Handler.Handle(ctx, r)
}

func traceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceIdFromHeader := r.Header["Traceid"]
		if len(traceIdFromHeader) != 0 {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), traceCtxKey, traceIdFromHeader[0])))
		} else {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), traceCtxKey, uuid.New().String())))
		}
	})
}

func contextMiddleware(ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rctx, cancel := context.WithCancel(r.Context())
		context.AfterFunc(ctx, cancel)
		next.ServeHTTP(w, r.WithContext(rctx))
	})
}
