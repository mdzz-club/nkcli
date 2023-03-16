package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	nkcli "github.com/mdzz-club/nkcli/internal"
	"github.com/urfave/cli/v2"
)

var (
	errInvalidRelayUrl = errors.New("Invalid relay URL")
)

func updateAction(c *cli.Context) error {
	dbpath := c.String("db")
	relays := c.StringSlice("relay")

	if relays == nil {
		relays = DefaultRelays
	}

	for _, url := range relays {
		if !strings.HasPrefix(url, "ws://") && !strings.HasPrefix(url, "wss://") {
			return errors.Join(errInvalidRelayUrl, errors.New(url))
		}
	}

	fmt.Print("Will use these relays to update your keys data:\n\n")

	for _, url := range relays {
		fmt.Printf("  %v\n", url)
	}

	db, err := nkcli.Open(dbpath)

	if err != nil {
		return err
	}

	defer db.Close()

	list, err := db.List()

	if err != nil {
		return err
	}

	fmt.Printf("\nFound %v keys\n\n", len(list))

	wg := new(sync.WaitGroup)

	ctx := context.WithValue(c.Context, "db", db)

	for _, item := range list {
		wg.Add(1)
		go nkcli.UpdateRelays(ctx, item, relays, wg)
	}

	wg.Wait()

	return nil
}
