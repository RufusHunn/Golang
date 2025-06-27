package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"text/template"
	"todos2/store"

	"github.com/google/uuid"
)

type contextKey int

const (
	traceCtxKey contextKey = iota + 1
)

var TraceIDString = uuid.New().String()

func getItem(w http.ResponseWriter, r *http.Request) {
	var fullItem = store.Get(r.PathValue("item"))
	slog.InfoContext(r.Context(), fmt.Sprintf("Found item: %s", fullItem))
}

func addItem(w http.ResponseWriter, r *http.Request) {
	var description = r.PathValue("item")
	var status = r.PathValue("status")
	fmt.Printf("Store lines are currently: %s", store.Lines)
	store.Create(description, status)
	slog.InfoContext(r.Context(), fmt.Sprintf("Added item: %s", description))
}

func updateItem(w http.ResponseWriter, r *http.Request) {
	var ix = r.PathValue("ix")
	var description = r.PathValue("item")
	var status = r.PathValue("status")
	fmt.Printf("Store lines are currently: %s", store.Lines)
	store.Update(ix, description, status)
	slog.InfoContext(r.Context(), fmt.Sprintf("Updated item: %s", description))
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	store.Delete(r.PathValue("item"))
	slog.InfoContext(r.Context(), fmt.Sprintf("Deleting item: %s", r.PathValue("item")))
}

func renderHTMLTable(w http.ResponseWriter, data map[string]store.Task) {
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
	tmpl.Execute(w, data)
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

	store.Load()

	fmt.Println("Lines loaded: ", store.Lines)

	slog.InfoContext(ctx, "loaded data")

	mux.HandleFunc("/get/{item}", getItem)
	mux.HandleFunc("/create/{item}/{status}", addItem)
	mux.HandleFunc("/update/{ix}/{item}/{status}", updateItem)
	mux.HandleFunc("/delete/{item}", deleteItem)

	// Below shows a navigable directory with one static page; is this sufficient?
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/about/", http.StripPrefix("/about", fs))

	// Dynamic webpage
	mux.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		renderHTMLTable(w, store.List())
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
	slog.InfoContext(ctx, "saving data")
	store.Save()
	fmt.Println("Exiting")

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
