package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/produce-logs", handleRequest)

	server := &http.Server{
		Addr:         ":80",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	var errChan = make(chan error, 1)
	go func() {
		_ = json.NewEncoder(os.Stdout).Encode(SLog{
			Timestamp: time.Now().Format(time.RFC3339Nano),
			Message:   "starting server on http://localhost:80",
		})

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Block until a signal is received
	var signalChan = make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signalChan:
		_ = json.NewEncoder(os.Stdout).Encode(SLog{
			Timestamp: time.Now().Format(time.RFC3339Nano),
			Message:   "received shutdown signal, shutting down gracefully",
		})

		// Attempt to shut down the server gracefully
		if err := server.Shutdown(context.Background()); err != nil {
			_ = json.NewEncoder(os.Stdout).Encode(SLog{
				Timestamp: time.Now().Format(time.RFC3339Nano),
				Message:   fmt.Sprintf("error shutting down server: %s", err),
			})
		}

		_ = json.NewEncoder(os.Stdout).Encode(SLog{
			Timestamp: time.Now().Format(time.RFC3339Nano),
			Message:   "server shutdown complete",
		})

	case err := <-errChan:
		if err != nil {
			_ = json.NewEncoder(os.Stdout).Encode(SLog{
				Timestamp: time.Now().Format(time.RFC3339Nano),
				Message:   fmt.Sprintf("error starting server: %s", err),
			})
		}
	}

}

type SLog struct {
	Timestamp  string `json:"timestamp"`
	Message    string `json:"message"`
	Attributes any    `json:"attributes,omitempty"`
}

func handleRequest(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var err error
	iter := 0
	iterationStr := strings.TrimSpace(r.URL.Query().Get("iteration"))

	logBytes, _ := json.Marshal(SLog{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Message:   "getting query params key: iteration",
		Attributes: map[string]string{
			"iteration": iterationStr,
		},
	})
	_, _ = fmt.Fprintln(os.Stdout, string(logBytes))

	if iterationStr != "" {
		iter, err = strconv.Atoi(iterationStr)
	}

	if err != nil {
		logData := SLog{
			Timestamp: time.Now().Format(time.RFC3339Nano),
			Message:   "cannot parse iteration",
			Attributes: map[string]string{
				"error": err.Error(),
			},
		}

		logBytes, _ = json.Marshal(logData)

		w.WriteHeader(http.StatusBadRequest)

		_, _ = fmt.Fprintln(os.Stdout, string(logBytes))
		_, _ = fmt.Fprintln(w, string(logBytes))
		return
	}

	wg := sync.WaitGroup{}
	for i := 0; i < iter; i++ {
		wg.Add(1)
		go func(idx int, wg *sync.WaitGroup) {
			defer wg.Done()

			time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)
			logData := SLog{
				Timestamp: time.Now().Format(time.RFC3339Nano),
				Message:   fmt.Sprintf("writing logs from worker %d", idx),
			}

			logBytes, _ = json.Marshal(logData)
			_, _ = fmt.Fprintln(os.Stdout, string(logBytes))
		}(i, &wg)

	}

	wg.Wait()

	logBytes, _ = json.Marshal(SLog{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Message:   fmt.Sprintf("successfully producting logs"),
		Attributes: map[string]interface{}{
			"iteration": iter,
		},
	})

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(os.Stdout, string(logBytes))
	_, _ = fmt.Fprintln(w, string(logBytes))
}
