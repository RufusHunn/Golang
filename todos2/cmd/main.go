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

var Tasks map[string]store.Task

type TaskMessage struct {
	Action   string
	Key      string
	Payload  store.Task
	Response chan interface{}
}

func generateID() string {

	keys := make([]int, 0, len(Tasks))
	for k := range Tasks {
		ik, _ := strconv.Atoi(k)
		keys = append(keys, ik)
	}
	nextKey := keys[len(keys)-1] + 1
	return fmt.Sprintf("%d", nextKey)
}

func getItem(ch chan TaskMessage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("ix")
		resp := make(chan interface{})
		ch <- TaskMessage{Action: "read", Key: id, Response: resp}
		result := <-resp

		if result == nil {
			http.NotFound(w, r)
			return
		}

		json.NewEncoder(w).Encode(result)
		slog.InfoContext(r.Context(), fmt.Sprintf("Retrieved item: %s", id))
	}
}

func addItem(ch chan TaskMessage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var description = r.PathValue("description")
		var status = r.PathValue("status")
		if description != "" && status != "" {
			var task = store.Task{Description: description, Status: status}
			id := generateID()
			resp := make(chan interface{})
			ch <- TaskMessage{Action: "create", Key: id, Payload: task, Response: resp}
			<-resp
			slog.InfoContext(r.Context(), fmt.Sprintf("Added item: %s", description))
		}
	}
}

func updateItem(ch chan TaskMessage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("ix")
		var task = store.Task{Description: r.PathValue("description"), Status: r.PathValue("status")}

		resp := make(chan interface{})
		ch <- TaskMessage{Action: "update", Key: id, Payload: task, Response: resp}
		result := <-resp

		if result == nil {
			http.NotFound(w, r)
			return
		}

		w.WriteHeader(http.StatusOK)
		slog.InfoContext(r.Context(), fmt.Sprintf("Updated item: %s", id))
	}
}

func deleteItem(ch chan TaskMessage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("ix")
		resp := make(chan interface{})
		ch <- TaskMessage{Action: "delete", Key: id, Response: resp}
		result := <-resp

		if result == nil {
			http.NotFound(w, r)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		slog.InfoContext(r.Context(), fmt.Sprintf("Deleted item: %s", id))
	}
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
	taskChan := make(chan TaskMessage)
	Tasks = store.Load()
	go taskActor(taskChan, Tasks)

	mux.HandleFunc("/get/{ix}", getItem(taskChan))
	mux.HandleFunc("/create/{description}/{status}", addItem(taskChan))
	mux.HandleFunc("/update/{ix}/{description}/{status}", updateItem(taskChan))
	mux.HandleFunc("/delete/{ix}", deleteItem(taskChan))

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

func taskActor(ch chan TaskMessage, tasks map[string]store.Task) {

	for msg := range ch {
		switch msg.Action {
		case "create":
			tasks[msg.Key] = msg.Payload
			msg.Response <- true
		case "read":
			task, ok := tasks[msg.Key]
			if ok {
				msg.Response <- task
			} else {
				msg.Response <- nil
			}
		case "update":
			_, exists := tasks[msg.Key]
			if exists {
				tasks[msg.Key] = msg.Payload
				msg.Response <- true
			} else {
				msg.Response <- nil
			}
		case "delete":
			_, exists := tasks[msg.Key]
			if exists {
				delete(tasks, msg.Key)
				msg.Response <- true
			} else {
				msg.Response <- nil
			}
		case "list":
			msg.Response <- tasks
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
