package chronos

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"sync"
	"text/template"
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

func (j *Job) generateTemplateFuncMap(env map[string]string) map[string]interface{} {
	return map[string]interface{}{
		"env": func(key string) string {
			value, ok := env[key]
			if ok {
				return value
			}
			return ""
		},
		"name": func() string {
			return j.name
		},
		"time": func(t string) string {
			return time.Now().Format(t)
		},
		"count": func() int {
			var count int
			for _, e := range j.execution {
				if e.succeeded {
					count++
				}
			}
			return count + 1
		},
	}
}

func (j *Job) generateEnvVariables(propagate bool) map[string]string {
	ret := make(map[string]string, len(j.task.Env))

	for name, value := range j.task.Env {
		ret[name] = value
	}

	if !propagate {
		return ret
	}

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		name := pair[0]
		value := pair[1]
		ret[name] = value
	}

	return ret
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
	if j.task.Timeout != 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(j.task.Timeout)*time.Second)
		defer cancel()
	}

	env := j.generateEnvVariables(j.task.PropagateEnv)
	args := make([]string, len(j.task.Args))
	copy(args, j.task.Args)
	if j.task.UseTemplate {
		tf := j.generateTemplateFuncMap(env)

		for i, arg := range args {
			var err error
			tmpl := template.Must(template.New("argtemplate").Funcs(tf).Parse(arg))
			if err != nil {
				return fmt.Errorf("failed to create template. templateText: %s, err: %w", arg, err)
			}

			w := new(bytes.Buffer)
			err = tmpl.Execute(w, nil)
			if err != nil {
				return fmt.Errorf("failed to apply template. templateText: %s, err: %w", arg, err)
			}

			j.logger.Debugf("transform args. before: %s, after:%s", args[i], w.String())
			args[i] = w.String()
		}
	}

	j.logger.Infof("Task %s started to execute command. command: %s %s", j.name, j.task.Command, strings.Join(j.task.Args, " "))
	cmd := exec.CommandContext(ctx, j.task.Command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if out := stdout.String(); len(out) != 0 {
		j.logger.Infof("Task %s outputted to STDOUT: %s", j.name, out)
	}
	if out := stderr.String(); len(out) != 0 {
		j.logger.Warnf("Task %s outputted to STDERR: %s", j.name, out)
	}
	if err != nil {
		j.logger.Warnf("Task %s failed to execute command: %s", j.name, err)
		return err
	}
	return nil
}

func ExecuteWithRetry(ctx context.Context, j *Job) {
	retryLimit := j.task.RetryLimit

	isRetryable := j.task.RetryLimit != RetryLimitNever
	isInfiniteRetry := j.task.RetryLimit == RetryLimitInfinite

	for i := 0; ; i++ {
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

		j.logger.Warnf("Task %s failed to execute command (failed %d of %d, will retry). err: %s", j.name, i, int(retryLimit), err)
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
		if !isInfiniteRetry && !isRetryable || i >= int(retryLimit) {
			break
		}

		time.Sleep(retryWait)
		i++
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
