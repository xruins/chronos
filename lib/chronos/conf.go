package chronos

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/BurntSushi/toml"
	validator "github.com/go-playground/validator/v10"
	yaml "gopkg.in/yaml.v3"
)

type Level string

const (
	LevelUnknown Level = ""
	LevelFatal   Level = "fatal"
	LevelError   Level = "error"
	LevelWarn    Level = "warn"
	LevelInfo    Level = "info"
	LevelDebug   Level = "debug"
)

type Config struct {
	LogLevel    Level            `validate:"oneof='fatal' 'error' 'warn' 'info' 'debug'|isdefault" json:"log_level" toml:"log_level" yaml:"log_level"`
	TimeZone    string           `validate:"timezone|isdefault" json:"time_zone" toml:"time_zone" yaml:"time_zone"`
	Tasks       map[string]*Task `validate:"required" json:"tasks" toml:"tasks" yaml:"tasks"`
	HealthCheck *HealthCheck     `json:"healthcheck" toml:"healthcheck" yaml:"healthcheck"`
}

func NewConfig(i io.Reader, filename string) (*Config, error) {
	conf := &Config{}
	extension := filepath.Ext(filename)

	b, err := ioutil.ReadAll(i)
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
		return nil, errors.New("the extension of config file must be one of yaml, yml, json and toml.")
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

type HealthCheck struct {
	Host string `validate:"required" json:"host" toml:"host" yaml:"host"`
	Port int    `validate:"gt=0,lte=65535" json:"port" toml:"port" yaml:"port"`
}

type RetryType string

const (
	RetryTypeFixed       RetryType = "fixed"
	RetryTypeExponential RetryType = "exponential"
)

type RetryLimit int

const (
	RetryLimitNever    RetryLimit = 0
	RetryLimitInfinite RetryLimit = -1
)

type Task struct {
	Description  string            `json:"description" json:"description" toml:"description" yaml:"description"`
	Command      string            `validate:"required" json:"command" toml:"command" yaml:"command"`
	Args         []string          `json:"args" toml:"args" yaml:"args"`
	Schedule     string            `validate:"required" json:"schedule" toml:"schedule" yaml:"schedule"`
	UseTemplate  bool              `json:"use_template" toml:"use_template" yaml:"use_template"`
	Env          map[string]string `json:"env" toml:"env" yaml:"env"`
	PropagateEnv bool              `json:"propagate_env" toml:"propagate_env" yaml:"propagate_env"`
	Timeout      int               `validate:"gte=0" json:"timeout" toml:"timeout" yaml:"timeout"`
	// RetryLimit is the count of retry to be attempted. 0: infinite, -1:never retry
	RetryLimit RetryLimit `validate:"gte=0" json:"retry_limit" toml:"retry_limit" yaml:"retry_limit"`
	// RetryWait is the time to wait before retry in second. With retry
	RetryWait int `validate:"gt=0" json:"retry_wait" toml:"retry_wait" yaml:"retry_wait"`
	// RetryType is the kind of retry. it must be one of `fixed` or `exponential`.
	// (fixed: retry with fixed wait time, exponential: retry with exponential backoff)
	RetryType    RetryType `validate:"oneof=fixed exponential" json:"retry_type" toml:"retry_type" yaml:"retry_type"`
	FailureCount int       `validate:"gte=0" json:"failure_count" toml:"failure_count" yaml:"failure_count"`
}
