package daemon

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const Port = 2187

var validEvents = map[string]bool{
	"prompt":             true,
	"thinking":           true,
	"stop":               true,
	"session_start":      true,
	"stop_failure":       true,
	"permission_request": true,
}

func ServeEvents(ctx context.Context, machine *Machine, debug bool) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /event", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 64))
		if err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}
		event := strings.TrimSpace(string(body))
		if !validEvents[event] {
			http.Error(w, "unknown event", http.StatusBadRequest)
			return
		}
		if debug {
			fmt.Printf("[%s] event: %s\n", time.Now().Format("2006-01-02 15:04:05.000"), event)
		}
		machine.Dispatch(event)
		w.WriteHeader(http.StatusNoContent)
	})

	server := &http.Server{Addr: fmt.Sprintf(":%d", Port), Handler: mux}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "claude2-d2: http server error: %v\n", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}
