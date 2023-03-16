package main

import (
	"encoding/hex"
	"fmt"

	nkcli "github.com/mdzz-club/nkcli/internal"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip06"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/tyler-smith/go-bip39"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
)

func generateAction(c *cli.Context) error {
	fmt.Print("Here's your mnemonic words:\n\n")

	ent, err := bip39.NewEntropy(128)

	if err != nil {
		return err
	}

	words, err := bip39.NewMnemonic(ent)

	if err != nil {
		return err
	}

	fmt.Printf("%v\n\n", words)

	priv, err := nip06.PrivateKeyFromSeed(nip06.SeedFromWords(words))

	if err != nil {
		return err
	}

	pub, err := nostr.GetPublicKey(priv)

	if err != nil {
		return err
	}

	fmt.Print("Enter a passphrase to protect your key:")

	password, err := terminal.ReadPassword(0)

	if err != nil {
		return err
	}

	keyBuf, err := hex.DecodeString(priv)

	if err != nil {
		return err
	}

	encKey, err := nkcli.Encrypt(keyBuf, password)

	dbpath := c.String("db")

	db, err := nkcli.Open(dbpath)

	if err != nil {
		return err
	}
	defer db.Close()

	err = db.SaveKey(pub, encKey)

	if err != nil {
		return err
	}

	bech32Pub, err := nip19.EncodePublicKey(pub)

	if err != nil {
		return err
	}

	fmt.Printf("\n\nYour public key:\n%v\n%v\n", pub, bech32Pub)

	return nil
}
