package main

import (
	"fmt"
	"os"

	"github.com/dkd/ccs/internal/cmd"
	"github.com/dkd/ccs/internal/display"
)

var version = "dev"

func main() {
	// Strip --json/--md global flags from args
	var filtered []string
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--json":
			display.OutputFormat = "json"
		case "--md":
			display.OutputFormat = "md"
		default:
			filtered = append(filtered, arg)
		}
	}
	os.Args = append(os.Args[:1], filtered...)

	command := "summary"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	var err error
	switch command {
	case "all":
		err = cmd.All()
	case "summary":
		err = cmd.Summary()
	case "today":
		err = cmd.Period("today")
	case "week":
		err = cmd.Period("week")
	case "month":
		err = cmd.Period("month")
	case "projects":
		err = cmd.Projects()
	case "sessions":
		err = cmd.Sessions(os.Args[2:])
	case "session":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: ccs session <id>")
			os.Exit(1)
		}
		err = cmd.SessionDetail(os.Args[2])
	case "tokens":
		err = cmd.Tokens()
	case "refresh":
		err = cmd.Refresh()
	case "version", "--version", "-v":
		fmt.Printf("ccs %s\n", version)
	case "help", "--help", "-h":
		cmd.Help(version)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		cmd.Help(version)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
