package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	nkcli "github.com/mdzz-club/nkcli/internal"
	"github.com/urfave/cli/v2"
)

func removeAction(c *cli.Context) error {
	db, err := nkcli.Open(c.String("db"))

	if err != nil {
		return err
	}

	if c.Args().Len() == 0 {
		return removeManually(db)
	} else {
		return removeKeys(db, c.Args().Slice())
	}
}

func removeManually(db *nkcli.DB) error {
	list, err := db.List()

	if err != nil {
		return err
	}

	if len(list) == 0 {
		fmt.Println("You don't have any keys, goto generate or import one.")
		return nil
	}

	fmt.Printf("You have %v keys:\n\n", len(list))

	nkcli.PrintKeyList(list)

	fmt.Print("\n  ⭐️ Choose one key: ")

	var line string
	_, err = fmt.Scanln(&line)

	if err != nil {
		return err
	}

	n, err := strconv.Atoi(line)

	if err != nil {
		return err
	}

	if n < 1 || n > len(list) {
		return errors.New("Invalid No.")
	}

	key := list[n-1]
	fmt.Printf("Do you want to DELETE '%v'? [y/n]", key.Pubkey)

	if nkcli.Scanline() != "y" {
		return nil
	}

	kb, err := hex.DecodeString(key.Pubkey)

	if err != nil {
		return err
	}

	return db.Remove(kb)
}

func removeKeys(db *nkcli.DB, keys []string) error {
	list := nkcli.SerializeKeys(keys)

	if len(list) == 0 {
		return nil
	}

	fmt.Print("Do you want to DELETE these keys?\n\n")

	for i, it := range list {
		fmt.Printf("  %v. %v\n", i+1, it)
	}

	fmt.Print("\n[y/n]")

	if nkcli.Scanline() != "y" {
		return nil
	}

	for _, k := range list {
		id, err := hex.DecodeString(k)

		if err != nil {
			return err
		}

		if db.Remove(id) != nil {
			return err
		}

		fmt.Printf("Key '%v' has been deleted", k)
	}

	return nil
}
