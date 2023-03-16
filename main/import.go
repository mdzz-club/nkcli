package main

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/mdzz-club/nkcli"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip06"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
)

func importAction(c *cli.Context) error {
	isRaw := c.Bool("raw")
	relays := c.StringSlice("relay")

	db, err := nkcli.Open(c.String("db"))

	if err != nil {
		return err
	}

	if relays == nil {
		relays = DefaultRelays
	}

	keys := make([]string, 0)
	if isRaw {
		if keys, err = importRawKeys(db, c.Args().Slice()); err != nil {
			return err
		}
	} else {
		if keys, err = importMnemonic(db, c.Args().Slice()); err != nil {
			return err
		}
	}

	if len(keys) == 0 {
		return nil
	}

	fmt.Print("\n\nNow update metadatas...\n\n")

	wg := new(sync.WaitGroup)
	ctx := context.WithValue(c.Context, "db", db)
	for _, k := range keys {
		info, err := db.GetKey(k, nil)

		if err != nil {
			continue
		}

		wg.Add(1)
		go nkcli.UpdateRelays(ctx, info, relays, wg)
	}

	wg.Wait()

	return nil
}

func importRawKeys(db *nkcli.DB, keys []string) (added []string, err error) {
	for _, sec := range nkcli.SerializeKeys(keys) {
		if db.Has(sec) {
			fmt.Printf("Skip, %v is exists", sec)
			continue
		}

		fmt.Print("Enter a passphrase to protect your key:")
		pass, err := terminal.ReadPassword(0)

		if err != nil {
			return nil, err
		}

		seckey, err := hex.DecodeString(sec)

		if err != nil {
			return nil, err
		}

		encKey, err := nkcli.Encrypt(seckey, pass)

		if err != nil {
			return nil, err
		}

		pub, err := nostr.GetPublicKey(sec)

		if err != nil {
			return nil, err
		}

		if db.Has(pub) {
			fmt.Printf("\n%v is exists, skip.\n", pub)
			continue
		}

		if err = db.SaveKey(pub, encKey); err != nil {
			return nil, err
		}

		fmt.Printf("\n%v saved.", pub)

		added = append(added, pub)
	}

	return
}

func importMnemonic(db *nkcli.DB, words []string) (keys []string, err error) {
	ws := strings.Join(words, " ")

	if !nip06.ValidateWords(ws) {
		err = errors.New("Invalid mnemonic words")
		return
	}

	seed := nip06.SeedFromWords(ws)
	priv, err := nip06.PrivateKeyFromSeed(seed)

	if err != nil {
		return
	}

	pub, err := nostr.GetPublicKey(priv)

	if err != nil {
		return nil, err
	}

	npub, err := nip19.EncodePublicKey(pub)

	if err != nil {
		return nil, err
	}

	fmt.Printf("\n\nThis is your public key: %v\n  Bech32 Encoded: %v\n\nIs corrent? [y/n]", pub, npub)

	if nkcli.Scanline() != "y" {
		return
	}

	if db.Has(pub) {
		fmt.Printf("\nThis key is already exists.\n")
		return
	}

	if err != nil {
		return
	}

	fmt.Print("\nEnter a password to protect your key: ")
	pass, err := terminal.ReadPassword(0)

	if err != nil {
		return
	}

	privBuf, err := hex.DecodeString(priv)

	if err != nil {
		return
	}

	enced, err := nkcli.Encrypt(privBuf, pass)

	if err != nil {
		return
	}

	db.SaveKey(pub, enced)

	fmt.Printf("\nYour key has been saved.")
	keys = append(keys, pub)

	return
}
