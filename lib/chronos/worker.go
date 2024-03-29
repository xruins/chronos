package chronos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/robfig/cron"
	"github.com/xruins/chronos/lib/logger"
)

// Worker is the implementation of Chronos worker.
type Worker struct {
	conf   *Config
	jobs   []*Job
	logger logger.Logger
	loc    *time.Location
}

// NewWorker returns an instance of `Worker`.
// It returns error when given malformed config.
func NewWorker(conf *Config, logger logger.Logger) (*Worker, error) {
	jobs := make([]*Job, 0, len(conf.Tasks))
	for name, t := range conf.Tasks {
		jobs = append(jobs, NewJob(name, t, logger))
	}

	loc := time.Local
	if tz := conf.TimeZone; tz != "" {
		var err error
		loc, err = time.LoadLocation(tz)
		if err != nil {
			return nil, fmt.Errorf("failed to get timezone: %w", err)
		}
	}

	return &Worker{
		conf:   conf,
		jobs:   jobs,
		logger: logger,
		loc:    loc,
	}, nil
}

type healthCheckResult struct {
	OK         bool     `json:"ok"`
	FailedJobs []string `json:"failed_jobs"`
}

func (w *Worker) healthCheckHandler(rw http.ResponseWriter, _ *http.Request) {
	var failedJobNames []string
	for _, j := range w.jobs {
		if !j.IsHealthy() {
			failedJobNames = append(failedJobNames, j.name)
		}
	}

	res := healthCheckResult{}
	if len(failedJobNames) > 0 {
		res.FailedJobs = failedJobNames
	} else {
		res.OK = true
	}

	b, err := json.Marshal(res)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write([]byte(fmt.Sprintf("failed to marshal JSON. err: %s", err)))
		return
	}

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(b)
	return
}

const (
	// HealthCheckEndpoint is an endpoint of health check API.
	HealthCheckEndpoint = "/health"
)

// ServeHealthCheckServer starts to serve HealthCheck server.
func (w *Worker) ServeHealthCheckServer() error {
	http.HandleFunc(HealthCheckEndpoint, w.healthCheckHandler)

	err := http.ListenAndServe(
		fmt.Sprintf("%s:%d", w.conf.HealthCheck.Host, w.conf.HealthCheck.Port),
		nil,
	)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("an error occured when serve HTTP server: %w", err)
	}
	return nil
}

// Run starts periodic execution of Jobs.
func (w *Worker) Run(ctx context.Context) error {
	stopCh := make(chan struct{}, 1)
	if w.conf.HealthCheck != nil {
		go func() {
			w.logger.Infof("Healthcheck server started on %s:%d", w.conf.HealthCheck.Host, w.conf.HealthCheck.Port)
			err := w.ServeHealthCheckServer()
			if !errors.Is(context.Canceled, err) {
				w.logger.Fatalf("Healthcheck server stopped. err: %s", err)
			}
			stopCh <- struct{}{}
		}()
	}

	c := cron.NewWithLocation(w.loc)
	w.logger.Infof("Worker started with timezone %s", w.loc)

	c.ErrorLog = log.Default()

	for _, j := range w.jobs {
		err := c.AddJob(j.task.Schedule, j)
		if err != nil {
			return fmt.Errorf("failed to add Task `%s`. err: %s", j.name, err)
		}
		w.logger.Infof("Task `%s` has been registered. schedule: %s", j.name, j.task.Schedule)
	}

	c.Start()
	defer c.Stop()

	for _, e := range c.Entries() {
		job, ok := e.Job.(*Job)
		if !ok {
			return fmt.Errorf("unexpected type of Job inside Entry. got: %T", e.Job)
		}

		w.logger.Infof("Task `%s` will be executed in %s at first", job.name, e.Next)
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("cancelled by context. err: %w", ctx.Err())
	case <-stopCh:
		return errors.New("cron scheduler stopped")
	}
}
