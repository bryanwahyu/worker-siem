package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/bryanwahyu/worker-siem/internal/interfaces/cli"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	args := os.Args[2:]
	ctx := context.Background()

	var err error

	switch subcommand {
	case "run":
		fs := flag.NewFlagSet("run", flag.ExitOnError)
		configPath := fs.String("config", "configs/worker.yaml", "Path to configuration file")
		dryRun := fs.Bool("dry-run", false, "Enable dry-run mode")
		_ = fs.Parse(args)
		err = cli.Run(ctx, *configPath, *dryRun)
	case "flush":
		fs := flag.NewFlagSet("flush", flag.ExitOnError)
		configPath := fs.String("config", "configs/worker.yaml", "Path to configuration file")
		_ = fs.Parse(args)
		err = cli.Flush(ctx, *configPath)
	case "test":
		fs := flag.NewFlagSet("test", flag.ExitOnError)
		configPath := fs.String("config", "configs/worker.yaml", "Path to configuration file")
		_ = fs.Parse(args)
		err = cli.Diagnostics(ctx, *configPath)
	case "help", "-h", "--help":
		printUsage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n", subcommand)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("mini-siem-worker <command> [options]")
	fmt.Println("commands:")
	fmt.Println("  run   - start the worker")
	fmt.Println("  flush - flush the spool")
	fmt.Println("  test  - run diagnostics")
}
