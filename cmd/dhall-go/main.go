package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"encoding/json"
	"github.com/urfave/cli/v2" // imports as package "cli"
	"github.com/wallyqs/dhall.go"
	"github.com/wallyqs/dhall.go/binary"
	"github.com/wallyqs/dhall.go/core"
	"github.com/wallyqs/dhall.go/imports"
	"github.com/wallyqs/dhall.go/parser"
	"gopkg.in/yaml.v2"
)

func main() {
	app := &cli.App{
		Name:  "dhall-golang",
		Usage: "Dhall implemented in Go",
		Commands: []*cli.Command{
			{
				Name:   "json",
				Usage:  "output Dhall code as JSON",
				Action: cmdJSON,
			},
			{
				Name:   "yaml",
				Usage:  "output Dhall code as YAML",
				Action: cmdYAML,
			},
		},
		Action: cmdDebug,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// cmdDebug is the original scrappy debug command
func cmdDebug(c *cli.Context) error {
	expr, err := parser.ParseReader("-", os.Stdin)
	if err != nil {
		return err
	}
	resolvedExpr, err := imports.Load(expr)
	if err != nil {
		return err
	}
	inferredType, err := core.TypeOf(resolvedExpr)
	if err != nil {
		return err
	}
	fmt.Fprint(os.Stderr, inferredType)
	fmt.Fprintln(os.Stderr)
	fmt.Println(core.Eval(resolvedExpr))

	var buf = new(bytes.Buffer)
	binary.EncodeAsCbor(buf, core.QuoteAlphaNormal(core.Eval(resolvedExpr)))
	final, err := binary.DecodeAsCbor(buf)
	if err != nil {
		return err
	}
	fmt.Printf("decoded as %+v\n", final)
	return nil
}

func cmdYAML(c *cli.Context) error {
	var data interface{}
	err := dhall.UnmarshalReader("-", os.Stdin, &data)
	if err != nil {
		return err
	}
	b, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Print(string(b))
	return nil
}

func cmdJSON(c *cli.Context) error {
	var data interface{}
	err := dhall.UnmarshalReader("-", os.Stdin, &data)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Print(string(b))
	return nil
}
