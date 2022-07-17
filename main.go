package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/xruins/chronos/lib/chronos"
	"github.com/xruins/chronos/lib/logger"
)

func init() {
	healthCheckCmd.PersistentFlags().IntP("timeout", "t", 30, "timeout to invoke healthcheck API in seconds")
}

var healthCheckCmd = &cobra.Command{
	Aliases: []string{"healthcheck"},
	Use:     "healthcheck",
	Example: "chronos healthcheck http://localhost:8080/healthcheck",
	Short:   "Invoke healthcheck API of Chronos worker",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Help()
			os.Exit(1)
		}

		timeout, err := cmd.Flags().GetInt("timeout")
		if err != nil {
			log.Fatalf("failed to get the value of `timeout` option: %s", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()
		u, err := url.Parse(args[0])
		if err != nil {
			log.Fatalf("failed to parse URL: %s", err)
		}
		client := chronos.NewClient(http.DefaultClient)
		ok, err := client.CheckHealth(ctx, u)
		if err != nil {
			log.Fatalf("failed to invoke healthcheck endpoint: %s", err)
		}
		if !ok {
			fmt.Fprintf(os.Stderr, "healthcheck API returned failed status")
			os.Exit(1)
		}
		fmt.Println("healthcheck OK")
		os.Exit(0)
	},
}

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start Chronos worker",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Help()
			os.Exit(1)
		}

		confName := args[0]
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
		w, err := chronos.NewWorker(conf, l)
		if err != nil {
			l.Fatalf("failed to start worker: %s", err)
		}
		ctx := context.Background()
		err = w.Run(ctx)
		if err != nil {
			l.Fatalf("failed to run worker: %s", err)
		}
		os.Exit(0)
	},
}

var rootCmd = &cobra.Command{
	Short: "chronos is implementation of the worker for periodic tasks",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(workerCmd, healthCheckCmd)
}

func main() {
	rootCmd.Execute()
}
