package main

import (
	"errors"
	"fmt"
	"strconv"

	nkcli "github.com/mdzz-club/nkcli/internal"
	"github.com/urfave/cli/v2"
)

func connectAction(c *cli.Context) error {
	if c.Args().Len() == 0 {
		fmt.Printf("You need pass a nostrconnect:// arg")
		return nil
	}

	cu, err := nkcli.ParseURL(c.Args().Get(0))

	if err != nil {
		return err
	}

	fmt.Printf("App name: %v\n", cu.Metadata.Name)
	if len(cu.Metadata.Url) != 0 {
		fmt.Printf("App URL: %v\n", cu.Metadata.Url)
	}
	if len(cu.Metadata.Description) != 0 {
		fmt.Printf("App description: %v\n", cu.Metadata.Description)
	}

	db, err := nkcli.Open(c.String("db"))

	if err != nil {
		return err
	}

	keys, err := db.List()

	if err != nil {
		return err
	}

	fmt.Printf("\nYou have %v keys:\n\n", len(keys))

	nkcli.PrintKeyList(keys)

	fmt.Print("  🔗 Choose your key: ")

	line := nkcli.Scanline()

	n, err := strconv.Atoi(line)

	if err != nil {
		return err
	}

	if n < 1 || n > len(keys) {
		return errors.New("Invalid key No.")
	}

	usedPub := keys[n-1]

	allows := []string{}

	if c.Bool("allow-all") {
		fmt.Println("\nNOTICE: This connection will allow all requests by default.")
		allows = []string{"get_public_key", "sign_event", "delegate", "get_relays", "nip04_encrypt", "nip04_decrypt"}
	}

	conn := &nkcli.Connection{
		AppID:  cu.Pubkey,
		Relay:  cu.Relay,
		PubKey: usedPub.Pubkey,
		Acked:  false,
		Allows: allows,
		Metadata: &nkcli.ConnMetadata{
			Name:        cu.Metadata.Name,
			Description: cu.Metadata.Description,
			Url:         cu.Metadata.Url,
		},
	}

	db.SetConnection(conn)
	fmt.Print("\nConnection info saved!\nRun nkcli without subcommand to serve it.\n")

	return nil
}
