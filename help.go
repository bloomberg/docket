package docket

import (
	"flag"
	"fmt"
	"os"
	"text/template"

	"github.com/fatih/color"
)

const docketHelp = `
Help for using docket:

  {{ var "GO_DOCKET_CONFIG" }}
    To use docket, set this to the name of the config to use.

Optional environment variables:

  {{ var "GO_DOCKET_DOWN" }} (default off)
      If non-empty, docket will run 'docker-compose down' after each suite.

  {{ var "GO_DOCKET_PULL" }} (default off)
      If non-empty, docket will run 'docker-compose pull' before each suite.
`

func init() {
	// We register a flag to get it shown in the default usage.
	//
	// We don't actually use the parsed flag value, though, since that would require us to call
	// flag.Parse() here. If we call flag.Parse(), then higher-level libraries can't easily add
	// their own flags, since testing's t.Run() will not re-run flag.Parse() if the flags have
	// already been parsed.
	//
	// Instead, we simply look for our flag text in os.Args.

	flag.Bool("help-docket", false, "get help on docket")

	for _, arg := range os.Args {
		if arg == "-help-docket" || arg == "--help-docket" {
			bold := func(s string) string {
				return color.New(color.Bold).Sprint(s)
			}

			tmpl := template.New("help").Funcs(template.FuncMap{"var": bold})
			template.Must(tmpl.Parse(docketHelp)).Execute(os.Stderr, nil)

			fmt.Fprintf(os.Stderr, "\n")
			os.Exit(2)
		}
	}
}
