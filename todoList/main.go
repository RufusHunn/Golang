package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"todoList/store"

	"github.com/google/uuid"
)

type contextKey int

const (
	traceCtxKey contextKey = iota + 1
)

func main() {

	var handler slog.Handler
	handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
	})

	handler = MyHandler{handler}

	slog.SetDefault(slog.New(handler))

	ctx := context.Background()
	ctx = context.WithValue(ctx, traceCtxKey, uuid.New().String())

	action := flag.String("action", "", "Action to perform")
	item := flag.String("item", "", "Item description")
	status := flag.String("status", "", "Item status")

	flag.Parse()

	fmt.Println("Action: ", *action)
	fmt.Println("Item: ", *item)
	fmt.Println("Status: ", *status)

	store.Load()

	switch *action {
	case "add":
		{
			store.Upsert(*item, *status)
			slog.InfoContext(ctx, fmt.Sprintf("upserting item: %s", *item))
		}
	case "delete":
		{
			store.Delete(*item)
			slog.InfoContext(ctx, fmt.Sprintf("deleting item: %s", *item))
		}
	case "list":
		{
			var allLines = store.List(*item)
			slog.InfoContext(ctx, fmt.Sprintf("listing items: %s", allLines))
		}
	case "get":
		{
			var fullItem = store.Get(*item)
			slog.InfoContext(ctx, fmt.Sprintf("Found item: %s", fullItem))
		}
	default:
		fmt.Println("Invalid action")
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	fmt.Println("App is running. Press Ctrl+C to exit...")
	<-sig

	store.Save()

	slog.InfoContext(ctx, "done")
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
