package main

import (
	"context"
	"log"
	"os"

	"github.com/xruins/chronos/lib/chronos"
	"github.com/xruins/chronos/lib/logger"
)

func main() {

	if len(os.Args) != 2 {
		log.Fatal("usage: chronos path-to-conf-file")
	}

	confName := os.Args[1]
	i, err := os.Open(confName)
	if err != nil {
		log.Fatalf("failed to open config file: %s", err)
	}
	conf, err := chronos.NewConfig(i, confName)
	if err != nil {
		log.Fatalf("failed to parse config file: %s", err)
	}

	l, err := logger.NewZapLogger(string(conf.LogLevel))
	defer l.Sync()
	if err != nil {
		log.Fatalf("failed to generate logger: %s", err)
	}
	w := chronos.NewWorker(conf, l)

	ctx := context.Background()
	err = w.Run(ctx)
	if err != nil {
		l.Fatalf("failed to run worker: %s", err)
	}
}
