package chronos_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/xruins/chronos/lib/chronos"
)

const (
	fixtureConfigJSON = `{
	"log_level": "debug",
        "time_zone": "Asia/Tokyo",
	"healthcheck": {
		"host": "0.0.0.0",
		"port": 30001
	},
	"tasks": {
		"hello": {
			"description": "hello",
			"command": "echo",
			"args": [
				"hoge",
				"fuga"
			],
                        "use_template": true,
			"schedule": "0 0 0 0 0",
			"env": {
				"TZ": "Asia/Tokyo"
			},
			"timeout": 30,
			"retry_limit": 3,
			"retry_wait": 30,
			"retry_type": "fixed",
			"failure_count": 1
		}
	}
}`

	fixtureConfigYAML = `
log_level: debug
time_zone: Asia/Tokyo
healthcheck:
  host: 0.0.0.0
  port: 30001
tasks:
  hello:
    description: hello
    command: echo
    args:
      - hoge
      - fuga
    use_template: true
    schedule: 0 0 0 0 0
    timeout: 30
    retry_limit: 3
    retry_wait: 30
    retry_type: fixed
    failure_count: 1
    env:
      TZ: Asia/Tokyo

`

	fixtureConfigTOML = `
log_level = "debug"
time_zone = "Asia/Tokyo"
[healthcheck]
host = "0.0.0.0"
port = 30001

[tasks.hello]
description = "hello"
command = "echo"
args = [ "hoge", "fuga" ]
schedule = "0 0 0 0 0"
use_template = true
timeout = 30
retry_limit = 3
retry_wait = 30
retry_type = "fixed"
failure_count = 1
  [tasks.hello.env]
  TZ = "Asia/Tokyo"
`
)

var fixtureConfig = &chronos.Config{
	LogLevel: "Debug",
	TimeZone: "Asia/Tokyo",
	HealthCheck: &chronos.HealthCheck{
		Host: "0.0.0.0",
		Port: 30001,
	},
	Tasks: map[string]*chronos.Task{
		"hello": &chronos.Task{
			Description: "hello",
			Command:     "echo",
			Args: []string{
				"hoge",
				"fuga",
			},
			UseTemplate:  true,
			Schedule:     "0 0 0 0 0",
			Timeout:      30,
			RetryLimit:   3,
			RetryWait:    30,
			RetryType:    chronos.RetryTypeFixed,
			FailureCount: 1,
			Env: map[string]string{
				"TZ": "Asia/Tokyo",
			},
		},
	},
}

func TestNewConfigWithJSON(t *testing.T) {
	r := strings.NewReader(fixtureConfigJSON)
	got, err := chronos.NewConfig(r, "test.json")
	if err != nil {
		t.Fatalf("failed to parse fixture JSON: %s", err)
	}

	want := fixtureConfig

	diff := cmp.Diff(want, got)
	if err != nil {
		t.Errorf("parsed config differs from the one expected: %s", diff)
	}
}

func TestNewConfigWithYAML(t *testing.T) {
	r := strings.NewReader(fixtureConfigYAML)
	got, err := chronos.NewConfig(r, "test.yml")
	if err != nil {
		t.Fatalf("failed to parse fixture YAML: %s", err)
	}

	want := fixtureConfig

	diff := cmp.Diff(want, got)
	if err != nil {
		t.Errorf("parsed config differs from the one expected: %s", diff)
	}
}

func TestNewConfigWithTOML(t *testing.T) {
	r := strings.NewReader(fixtureConfigTOML)
	got, err := chronos.NewConfig(r, "test.toml")
	if err != nil {
		t.Fatalf("failed to parse fixture TOML: %s", err)
	}

	want := fixtureConfig

	diff := cmp.Diff(want, got)
	if err != nil {
		t.Errorf("parsed config differs from the one expected: %s", diff)
	}
}
