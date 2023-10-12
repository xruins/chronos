package chronos

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// Level is the log level for the logger.
type Level string

const (
	// LevelUnknown is the default value of `Level`.
	// It is same to specify `info` level.
	LevelUnknown Level = ""
	// LevelFatal is the log level to show logs on `fatal` and above.
	LevelFatal Level = "fatal"
	// LevelError is the log level to show logs on `error` and above.
	LevelError Level = "error"
	// LevelWarn is the log level to show logs on `warn` and above.
	LevelWarn Level = "warn"
	// LevelInfo is the log level to show logs on `info` and above.
	LevelInfo Level = "info"
	// LevelDebug is the log level to show logs on `debug` and above.
	// This level is not recommended for regular use because it outputs many verbose logs.
	LevelDebug Level = "debug"
)

// Config is the config for Chronos process.
type Config struct {
	// LogLevel is the level for logging. it must be one of 'fatal', 'error', 'warn', 'info' and 'debug'.
	// By default, use `info` level.
	LogLevel Level `validate:"oneof='fatal' 'error' 'warn' 'info' 'debug'|isdefault" json:"log_level" toml:"log_level" yaml:"log_level"`
	// TimeZone is a time-zone which applied to execution time of tasks. By default, use `Local`.
	TimeZone string `validate:"timezone|isdefault" json:"time_zone" toml:"time_zone" yaml:"time_zone"`
	// Tasks are the task which executed periodically.
	Tasks map[string]*Task `validate:"required" json:"tasks" toml:"tasks" yaml:"tasks"`
	// HealthCheck is the settings for HealthCheck API.
	HealthCheck *HealthCheck `json:"healthcheck" toml:"healthcheck" yaml:"healthcheck"`
}

// NewConfig return the instance of Config.
// It returns error when failed to read the file or read malformed config.
func NewConfig(i io.Reader, filename string) (*Config, error) {
	conf := &Config{}
	extension := filepath.Ext(filename)

	b, err := io.ReadAll(i)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	switch extension {
	case ".yml", ".yaml":
		err = yaml.Unmarshal(b, conf)
	case ".json":
		err = json.Unmarshal(b, conf)
	case ".toml":
		err = toml.Unmarshal(b, conf)
	default:
		return nil, errors.New("the extension of config file must be one of yaml, yml, json and toml")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	validate := validator.New()
	err = validate.Struct(conf)
	if err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return conf, nil
}

// HealthCheck is the configuration for HealthCheck server.
type HealthCheck struct {
	// Host is the host to bind by HealthCheck server. By default, use `localhost`.
	Host string `validate:"required" json:"host" toml:"host" yaml:"host"`
	// Port is the TCP port to be used by HealthCheck server, By default, use 8080.
	Port int `validate:"gt=0,lte=65535" json:"port" toml:"port" yaml:"port"`
}

// RetryType is the enum of the ways of command retry.
type RetryType string

const (
	// RetryTypeFixed is the kind of retry, which retries on fixed interval.
	RetryTypeFixed RetryType = "fixed"
	// RetryTypeExponential is the kind of retry, which retries on exponential backoff.
	RetryTypeExponential RetryType = "exponential"
)

// RetryLimit is the number to limit how many times to attempt retry.
type RetryLimit int

const (
	// RetryLimitNever is the number not to attempt retry.
	RetryLimitNever RetryLimit = 0
	// RetryLimitInfinite is the number to attempt retry without limit.
	// This value is used by default.
	RetryLimitInfinite RetryLimit = -1
)

type Task struct {
	// Description is a description of task.
	Description string `json:"description" json:"description" toml:"description" yaml:"description"`
	// Command is the executable name to exec.
	Command string `validate:"required" json:"command" toml:"command" yaml:"command"`
	// Args are the argument given for `Command`.
	Args []string `json:"args" toml:"args" yaml:"args"`
	// Schedule is the specification of the interval of task execution.
	// [examples]
	// `0 0 * * * *` (Every hour on the half hour) (Seconds, Minutes, Hours, Day of month, Month, Day of week)
	// `@hourly` (Every hour)
	// `@every 2h15m` (Every two hour fifteen
	Schedule string `validate:"required" json:"schedule" toml:"schedule" yaml:"schedule"`
	// UseTemplate is the option to enable template for `Args`.
	// If true, the following templates are available on `Args`.
	// `{{env "env_name"}}`: replaced with ENV["env_name"].
	// `{{time "2006-01-02T15:04:05Z07:00"}}: replaced with the current time formed as `2020-01-01T00:00:00Z07:00`.
	// see https://pkg.go.dev/time#pkg-constants for time format.
	// `{{count}}`: replaced with the times of successful executions.
	UseTemplate bool `json:"use_template" toml:"use_template" yaml:"use_template"`
	// Env is the environment variables which given for command.
	Env map[string]string `json:"env" toml:"env" yaml:"env"`
	// PropagateEnv is the switch to enable propagation of environment values.
	// If true, the environment variables given for Chronos worker are applied for command.
	PropagateEnv bool `json:"propagate_env" toml:"propagate_env" yaml:"propagate_env"`
	// Timeout is the seconds for timeout of command.
	Timeout int `validate:"gte=0" json:"timeout" toml:"timeout" yaml:"timeout"`
	// RetryLimit is the count of retry to be attempted. 0: infinite, -1:never retry
	RetryLimit RetryLimit `validate:"gte=0" json:"retry_limit" toml:"retry_limit" yaml:"retry_limit"`
	// RetryWait is the time to wait before retry in second.
	RetryWait int `validate:"gt=0" json:"retry_wait" toml:"retry_wait" yaml:"retry_wait"`
	// RetryType is the kind of retry. it must be one of `fixed` or `exponential`.
	// (fixed: retry with fixed wait time, exponential: retry with exponential backoff)
	RetryType RetryType `validate:"oneof=fixed exponential" json:"retry_type" toml:"retry_type" yaml:"retry_type"`
	// FailureCount is the number of failure which makes HealthCheck failed.
	// If the command failed `FailureCount` times or more, HealthCheck for the task shows failing status.
	FailureCount int `validate:"gte=0" json:"failure_count" toml:"failure_count" yaml:"failure_count"`
}
