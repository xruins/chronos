package chronos

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/k0kubun/pp"
	"github.com/robfig/cron"
	"github.com/xruins/chronos/lib/logger"
)

type Worker struct {
	conf   *Config
	jobs   []*Job
	logger logger.Logger
}

func NewWorker(conf *Config, logger logger.Logger) *Worker {
	jobs := make([]*Job, 0, len(conf.Tasks))
	for name, t := range conf.Tasks {
		jobs = append(jobs, NewJob(name, t, logger))
	}

	return &Worker{
		conf:   conf,
		jobs:   jobs,
		logger: logger,
	}
}

func (w *Worker) healthCheckHandler(rw http.ResponseWriter, r *http.Request) {
	for _, j := range w.jobs {
		if !j.IsHealthy() {
			rw.WriteHeader(http.StatusPreconditionFailed)
			return
		}
	}
	rw.WriteHeader(http.StatusOK)
}

func (w *Worker) ServeHealthCheckServer() error {
	http.HandleFunc("/health", w.healthCheckHandler)
	err := http.ListenAndServe(
		fmt.Sprintf("%s:%d", w.conf.HealthCheck.Host, w.conf.HealthCheck.Port),
		nil,
	)
	if err != nil {
		return fmt.Errorf("an error occured when serve HTTP server: %w", err)
	}
	return nil
}

func (w *Worker) Run(ctx context.Context) error {
	stopCh := make(chan struct{}, 1)
	if w.conf.HealthCheck != nil {
		go func() {
			w.logger.Infof("healthcheck server started on %s:%d", w.conf.HealthCheck.Host, w.conf.HealthCheck.Port)
			err := w.ServeHealthCheckServer()
			if !errors.Is(context.Canceled, err) {
				w.logger.Fatalf("healthcheck server stopped. err: %s", err)
			}
			stopCh <- struct{}{}
		}()
	}

	var c *cron.Cron
	if tz := w.conf.TimeZone; tz != "" {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			return fmt.Errorf("failed to get timezone: %w", err)
		}
		w.logger.Info("Worker started with timezone %s", loc)
		c = cron.NewWithLocation(loc)
	} else {
		c = cron.New()
	}

	c.ErrorLog = log.Default()

	for i, j := range w.jobs {
		w.logger.Debugf("jobs[%d]: %s", i, pp.Sprintf("%s", j))
	}

	for _, j := range w.jobs {
		w.logger.Debugf("tasks: %s", *j)
		c.AddFunc(
			j.task.Schedule,
			func() {
				ExecuteWithRetry(ctx, j)
			},
		)
		w.logger.Infof("Task %s has been registered. schedule %s", j.name, j.task.Schedule)
	}

	for i, e := range c.Entries() {
		w.logger.Debugf("entries[%d]: %s", i, pp.Sprintf("%s", e))
	}

	c.Start()
	defer c.Stop()

	select {
	case <-ctx.Done():
		return fmt.Errorf("cancelled by context. err: %w", ctx.Err())
	case <-stopCh:
		return errors.New("cron scheduler stopped")
	}

}
