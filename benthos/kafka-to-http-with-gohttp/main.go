package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type Handler struct{}

// ServeHTTP will handle any request and return http_status based on query param (by default 422)
// Also, will sleep_for if the parameter is exist in query param (by default 30s).
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	type Response struct {
		Message        string      `json:"message,omitempty"`
		Path           string      `json:"path,omitempty"`
		Method         string      `json:"method,omitempty"`
		Host           string      `json:"host,omitempty"`
		UserAgent      string      `json:"userAgent,omitempty"`
		ReqHeader      http.Header `json:"reqHeader,omitempty"`
		ReqQueryParams url.Values  `json:"reqQueryParams,omitempty"`
		ReqRawBody     string      `json:"reqRawBody,omitempty"`
	}

	log.Printf("[%s] %s\n", r.Method, r.URL.String())

	// copy body bytes so it can be re-readed
	var bodyBytes []byte
	var err error
	if r.Body != nil {
		bodyBytes, err = io.ReadAll(r.Body)
		if err == nil {
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	// parse form for query param
	err = r.ParseForm()
	if err != nil {
		err = fmt.Errorf("cannot do ParseForm: %w", err)

		resp, _ := json.Marshal(Response{
			Message:        fmt.Sprintf("%s", err),
			Path:           r.URL.Path,
			Method:         r.Method,
			Host:           r.Host,
			UserAgent:      r.UserAgent(),
			ReqHeader:      r.Header,
			ReqQueryParams: r.URL.Query(),
			ReqRawBody:     string(bodyBytes),
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write(resp)
		return
	}

	msg := "no status"
	httpStatus, _ := strconv.Atoi(r.URL.Query().Get("http_status"))
	if httpStatus == 0 {
		httpStatus = http.StatusExpectationFailed
		msg = fmt.Sprintf("http status %d [failed decode]", httpStatus)
	} else {
		msg = fmt.Sprintf("http status %d [by user]", httpStatus)
	}

	sleepFor, err := time.ParseDuration(r.URL.Query().Get("sleep_for"))
	if err != nil {
		sleepFor = 30 * time.Second
		msg = fmt.Sprintf("%s; timeout %s [default]", msg, sleepFor)
	} else {
		msg = fmt.Sprintf("%s; timeout %s [by user]", msg, sleepFor)
	}

	// take a nap
	time.Sleep(sleepFor)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	resp, _ := json.Marshal(Response{
		Message:        msg,
		Path:           r.URL.Path,
		Method:         r.Method,
		Host:           r.Host,
		UserAgent:      r.UserAgent(),
		ReqHeader:      r.Header,
		ReqQueryParams: r.URL.Query(),
		ReqRawBody:     string(bodyBytes),
	})
	_, _ = w.Write(resp)
}

func main() {
	var h Handler

	ctx := context.Background()
	httpServer := &http.Server{
		Addr:    ":4000",
		Handler: h,
	}

	var apiErrChan = make(chan error, 1)
	go func() {
		log.Printf("server start at port %s\n", httpServer.Addr)
		err := httpServer.ListenAndServe()
		if err != nil {
			apiErrChan <- fmt.Errorf("cannot start server")
		}

	}()
	// ** listen for sigterm signal
	var signalChan = make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signalChan:
		log.Println(ctx, "system: exiting...")
		log.Println(ctx, "http transport: exiting...")
		if _err := httpServer.Shutdown(ctx); _err != nil {
			log.Fatalf("http transport error: %s\n", _err)
		}

	case err := <-apiErrChan:
		if err != nil {
			log.Fatalf("http server error: %s\n", err)
		}
	}

}
