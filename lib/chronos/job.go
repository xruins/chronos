package chronos

import (
	"bytes"
	"context"
	"math"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/xruins/chronos/lib/logger"
)

type State int

const (
	StateUnknown = iota
	StateHealthy
	StateUnhealthy
)

type Job struct {
	name       string
	mu         sync.RWMutex
	task       *Task
	retryCount int
	State      State
	execution  []*Execution
	logger     logger.Logger
}

func NewJob(name string, task *Task, logger logger.Logger) *Job {
	return &Job{
		name:   name,
		task:   task,
		mu:     sync.RWMutex{},
		State:  StateHealthy,
		logger: logger,
	}
}

func (j *Job) IsHealthy() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.State == StateHealthy
}

func (j *Job) Execute(ctx context.Context) error {
	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, time.Duration(j.task.Timeout)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, j.task.Command, j.task.Args...)
	env := os.Environ()
	for envName, envValue := range j.task.Env {
		env = append(env, envName+"="+envValue)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		j.logger.Warnf("Task %s failed to execute command: %s", err)
		return err
	}
	if out := stdout.String(); len(out) != 0 {
		j.logger.Infof("Task %s outputted to STDOUT: %s", j.name, out)
	}
	if out := stderr.String(); len(out) != 0 {
		j.logger.Warnf("Task %s outputted to STDERR: %s", j.name, out)
	}
	return nil
}

func ExecuteWithRetry(ctx context.Context, j *Job) {
	j.logger.Infof("Task %s started to execute command. command: %s %s", j.name, j.task.Command, strings.Join(j.task.Args, " "))
	retryLimit := j.task.RetryLimit

	isRetryable := j.task.RetryLimit != RetryLimitNever
	isInfiniteRetry := j.task.RetryLimit == RetryLimitInfinite

	for i := 0; isInfiniteRetry || isRetryable && i < int(retryLimit); i++ {
		execution := &Execution{
			count:        i,
			executedTime: time.Now(),
		}
		err := j.Execute(ctx)
		if err == nil {
			execution.succeeded = true
			j.logger.Infof("Task %s finished to execute command successfully.", j.name)
			j.execution = append(j.execution, execution)
			return
		}

		j.logger.Warnf("Task %s failed to execute command (will retry).", j.name)
		execution.err = err
		j.execution = append(j.execution, execution)

		var retryWait time.Duration
		switch j.task.RetryType {
		case RetryTypeFixed:
			retryWait = time.Duration(j.task.RetryWait) * time.Second
		case RetryTypeExponential:
			retryWait = time.Duration(int(math.Pow(2, float64(i)))*j.task.RetryWait) * time.Second
		default:
			j.logger.Errorf("malformed retry_type for Task %s. retry_type: %s", j.name, j.task.RetryType)
			break
		}
		time.Sleep(retryWait)
	}

	j.logger.Errorf("Task %s exceeded to retry limit.", j.name)
	j.mu.Lock()
	j.State = StateUnhealthy
	j.mu.Unlock()
	return
}

type Execution struct {
	count        int
	executedTime time.Time
	err          error
	succeeded    bool
}
