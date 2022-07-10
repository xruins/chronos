package chronos_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/xruins/chronos/lib/chronos"
)

func TestWorkerHealthCheck(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := chronos.NewWorker(fixtureConfig)
	go func(ctx context.Context) {
		w.ServeHealthCheckServer()
	}(ctx)

	url := fmt.Sprintf("http://%s:%d/healthcheck", fixtureConfig.HealthCheck.Host, fixtureConfig.HealthCheck.Port)
	r, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to request for healthcheck end-point: %s", err)
	}

	got := r.StatusCode
	want := http.StatusOK
	if want != got {
		t.Errorf("unexpected http status code. got: %d, want: %d", got, want)
	}
}
