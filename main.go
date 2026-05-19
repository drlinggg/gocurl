package main

import (
	"fmt"
	"os"

	cmdio "github.com/drlinggg/gocurl/cmd/gocurl/io"
	"github.com/drlinggg/gocurl/cmd/gocurl/logger"
	"github.com/drlinggg/gocurl/core"
	corehttp "github.com/drlinggg/gocurl/core/http"
	"github.com/drlinggg/gocurl/storage"
)

const helpText = `gocurl — curl replacement

USAGE
  gocurl [flags] [METHOD] <url|preset[/path]> [fields...]
  gocurl set <preset> <field=value> [field=value ...]

METHOD
  GET POST PUT PATCH DELETE OPTIONS HEAD  (default: GET, or POST if body given)

URL
  example.com/api          plain host (scheme from preset/default)
  dev /users               preset name + path → expands to preset base + /users

FLAGS
  --pretty                 pretty-print JSON response
  -v                       show response headers
  --http1|--http2|--http3  force HTTP version
  --scheme http|https
  -t <seconds>             request timeout
  --help                   show this help

FIELDS
  @Key=Value               request header
  ?key=value               query parameter
  key=value                body field (JSON string)
  key:=value               body field (JSON raw value)

SET COMMAND
  gocurl set <preset> base=https://api.dev.com
  gocurl set <preset> @Authorization=Bearer $TOKEN
  gocurl set <preset> ?api_key=abc
  gocurl set <preset> timeout=30
  gocurl set <preset> scheme=https
  gocurl set <preset> http=2

EXAMPLES
  gocurl httpbin.org/get
  gocurl --pretty POST httpbin.org/post name=alex age:=30
  gocurl set dev base=https://api.dev.com
  gocurl dev /users
`

func main() {
	args := os.Args[1:]

	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		fmt.Print(helpText)
		return
	}

	log := logger.New()
	_ = log

	cfg, err := storage.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if args[0] == "set" {
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: gocurl set <preset> <field=value> ...")
			os.Exit(1)
		}
		presetName := args[1]
		for _, arg := range args[2:] {
			if err := cfg.SetField(presetName, arg); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		if err := cfg.Save(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("preset %q updated\n", presetName)
		return
	}

	hist, err := storage.LoadHistory()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	engine := core.New(
		&corehttp.Sender{},
		cfg, hist,
		cmdio.NewInput(),
		cmdio.NewOutput(os.Stdout, cfg.Colors()),
	)

	if err := engine.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
