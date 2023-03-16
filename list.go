package main

import (
	"fmt"

	nkcli "github.com/mdzz-club/nkcli/internal"
	"github.com/urfave/cli/v2"
)

func listAction(c *cli.Context) error {
	dbpath := c.String("db")

	db, err := nkcli.Open(dbpath)

	if err != nil {
		return err
	}

	defer db.Close()

	list, err := db.List()

	if err != nil {
		return err
	}

	if len(list) == 0 {
		fmt.Print("You don't have any keys, generate one or import.\n")
		return nil
	}

	fmt.Printf("You have %v keys:\n\n", len(list))

	nkcli.PrintKeyList(list)

	return nil
}
