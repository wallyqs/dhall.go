package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/wallyqs/dhall.go"
	"gopkg.in/yaml.v2"
)

const version = "0.1.0"

func showVersionAndExit() {
	fmt.Printf("dhall-go v%s\n", version)
	os.Exit(0)
}

type config struct {
	showVersion  bool
	showHelp     bool
	outputFormat string
	file         string
}

const helpText = `dhall-go

Examples:
  # Output as YAML
  dhall-go -f file.dhall -o yaml

  # Output as JSON
  dhall-go -f file.dhall -o json

Global Flags:
  -h, --help                    Show context-sensitive help.
      --version                 Show application version.
`

func main() {
	fs := flag.NewFlagSet("dhall-go", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Printf("usage: dhall-go\n")
		fs.PrintDefaults()
		fmt.Println()
	}
	cfg := &config{}
	fs.BoolVar(&cfg.showHelp, "h", false, "Show help")
	fs.BoolVar(&cfg.showHelp, "help", false, "Show help")
	fs.BoolVar(&cfg.showVersion, "version", false, "Show version")
	fs.BoolVar(&cfg.showVersion, "v", false, "Show version")
	fs.StringVar(&cfg.file, "f", "", "Configuration file")
	fs.StringVar(&cfg.file, "file", "", "Configuration file")
	fs.StringVar(&cfg.outputFormat, "o", "yaml", "Output format (yaml, json)")
	fs.StringVar(&cfg.outputFormat, "output", "yaml", "Output format (yaml, json)")
	fs.Parse(os.Args[1:])

	if cfg.showHelp {
		fmt.Fprintln(os.Stderr, helpText)
		os.Exit(0)
	}

	if cfg.showVersion {
		showVersionAndExit()
	}

	var data interface{}
	err := dhall.UnmarshalFile(cfg.file, &data)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("dhall-go: %w", err))
		os.Exit(1)
	}

	var b []byte
	switch cfg.outputFormat {
	case "yaml":
		b, err = yaml.Marshal(data)
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Errorf("dhall-go: %w", err))
			os.Exit(1)
		}
	case "json":
		b, err = json.MarshalIndent(data, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Errorf("dhall-go: %w", err))
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, fmt.Errorf("dhall-go: undefined format %s", cfg.outputFormat))
		os.Exit(1)
	}
	fmt.Println(string(b))
}
