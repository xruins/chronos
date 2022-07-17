package chronos_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/xruins/chronos/lib/chronos"
	"github.com/xruins/chronos/lib/logger"
)

func TestWorkerHealthCheck(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	w, err := chronos.NewWorker(fixtureConfig, &logger.NopLogger{})
	if err != nil {
		t.Fatalf("failed to creare worker: %s", err)
	}
	go func(ctx context.Context) {
		w.ServeHealthCheckServer()
	}(ctx)
	time.Sleep(100 * time.Millisecond)

	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/health", fixtureConfig.HealthCheck.Port), nil)
	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}
	req.WithContext(ctx)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to request for healthcheck end-point: %s", err)
	}
	res.Body.Close()

	got := res.StatusCode
	want := http.StatusOK
	if want != got {
		t.Errorf("unexpected http status code. got: %d, want: %d", got, want)
	}
}

func TestWorkerRun(t *testing.T) {
	t.Skip()
	counterCh1 := make(chan struct{}, 1)
	counterCh2 := make(chan struct{}, 1)

	mux := http.NewServeMux()
	mux.Handle("/test1", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Println("test1")
		counterCh1 <- struct{}{}
		w.WriteHeader(http.StatusOK)
	}))
	mux.Handle("/test2", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Println("test2")
		counterCh2 <- struct{}{}
		w.WriteHeader(http.StatusOK)
	}))
	go func() {
		if err := http.ListenAndServe(":30000", mux); !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("test server exited with error: %s", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	conf := &chronos.Config{
		Tasks: map[string]*chronos.Task{
			"test1": &chronos.Task{
				Command:    "curl",
				Args:       []string{"-fsSL", "http://localhost:30000/test1"},
				Schedule:   "@every 1s",
				RetryLimit: -1, // infinite
				RetryType:  "fixed",
			},
			"test2": &chronos.Task{
				Command:    "curl",
				Args:       []string{"-fsSL", "http://localhost:30000/test2"},
				Schedule:   "@every 1s",
				RetryLimit: -1, // infinite
				RetryType:  "fixed",
			},
		},
	}

	l, err := logger.NewZapLogger("debug")
	if err != nil {
		t.Fatalf("failed to create logger: %s", err)
	}
	w, err := chronos.NewWorker(conf, l)
	if err != nil {
		t.Fatalf("failed to creare worker: %s", err)
	}
	go func(ctx context.Context) {
		err := w.Run(ctx)
		if err != nil {
			t.Fatalf("worker exited with error: %s", err)
		}
	}(ctx)

	var isTest1OK, isTest2OK bool
	for {
		select {
		case <-ctx.Done():
			t.Fatal("timed out (cancelled without receiving channel messages)")
		case <-counterCh1:
			isTest1OK = true
		case <-counterCh2:
			isTest2OK = true
		default:
		}
		if isTest1OK && isTest2OK {
			break
		}
	}
}
