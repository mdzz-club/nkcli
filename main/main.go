package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

var (
	version = "0.0.0"
)

var (
	DefaultRelays = []string{
		"wss://relay.damus.io",
		"wss://nostring.deno.dev",
		"wss://eden.nostr.land",
		"wss://relay.nostr.band",
	}
)

func main() {
	dbpath, err := getDBPath()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	app := &cli.App{
		Name:  "nkcli",
		Usage: "Manage Nostr keys",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "db",
				Aliases: []string{"d"},
				Usage:   "Database file",
				Value:   dbpath,
				EnvVars: []string{"NKCLI_DB"},
			},
		},
		Action:  serveAction,
		Version: version,
		Commands: []*cli.Command{
			{
				Name:    "generate",
				Aliases: []string{"g"},
				Usage:   "Generate a new key",
				Action:  generateAction,
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "List keys",
				Action:  listAction,
			},
			{
				Name:    "update",
				Aliases: []string{"u"},
				Usage:   "Update keys metadata and relay list",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "relay",
						Aliases: []string{"r"},
						Usage:   "Use specific relay",
					},
				},
				Action: updateAction,
			},
			{
				Name:    "import",
				Aliases: []string{"i"},
				Usage:   "Import your key",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "relay",
						Aliases: []string{"r"},
						Usage:   "Use specific relay to retrieve metadata",
					},
					&cli.BoolFlag{
						Name:  "raw",
						Usage: "Use raw nsec1 or hex encoded private key",
						Value: false,
					},
				},
				Action: importAction,
			},
			{
				Name:   "remove",
				Usage:  "Remove your key and connected sessions",
				Action: removeAction,
			},
			{
				Name:    "connect",
				Aliases: []string{"c"},
				Usage:   "Create new connection via nostrconnect://",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "allow-all",
						Aliases: []string{"A"},
						Usage:   "Allow all request always",
						Value:   false,
					},
				},
				ArgsUsage: "nostrconnect://...",
				Action:    connectAction,
			},
			{
				Name:   "disconnect",
				Usage:  "Disconnect and remove connection",
				Action: disconnectAction,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}

func getDBPath() (string, error) {
	dir, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	return dir + "/.nkclidb", nil
}
