package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/mdzz-club/nkcli"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
)

func disconnectAction(c *cli.Context) error {
	dbpath := c.String("db")
	db, err := nkcli.Open(dbpath)

	if err != nil {
		return err
	}

	conns, err := db.ListConnection()

	if err != nil {
		return err
	}

	if len(conns) == 0 {
		fmt.Println("You don't have any connections.")
		return nil
	}

	fmt.Printf("You have %v connections:\n\n", len(conns))

	for i, c := range conns {
		fmt.Printf("%v. %v\n", i+1, c.Metadata.Name)
	}

	fmt.Print("\n  Choose one to disconnect ✂️ : ")

	line := nkcli.Scanline()

	if len(line) == 0 {
		return nil
	}

	n, err := strconv.Atoi(line)

	if err != nil {
		return err
	}

	if n < 1 || n > len(conns) {
		return errors.New("Invalid No. Out of range.")
	}

	conn := conns[n-1]

	if conn.KeyInfo == nil {
		fmt.Print("Enter your passphrase to unlock your private key:")
		pass, err := terminal.ReadPassword(0)

		if err != nil {
			return err
		}

		info, err := db.GetKey(conn.PubKey, pass)

		if err != nil {
			return err
		}

		conn.KeyInfo = info
	}

	err = sendDisconnect(c.Context, conn)

	if err != nil {
		return err
	}

	db.Disconnect(conn.AppID)

	fmt.Println("\nYour connection has been disconnected.")

	return nil
}

func sendDisconnect(ctx context.Context, conn *nkcli.Connection) error {
	relay, err := nostr.RelayConnect(ctx, conn.Relay)

	if err != nil {
		return err
	}

	rb := make([]byte, 16)
	rand.Read(rb)
	randomId := hex.EncodeToString(rb)
	data := map[string]any{"id": randomId, "method": "disconnect", "params": []string{}}

	shared, err := nip04.ComputeSharedSecret(conn.AppID, conn.KeyInfo.Privkey)

	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	jstr, err := json.Marshal(data)

	if err != nil {
		return err
	}

	content, err := nip04.Encrypt(string(jstr), shared)
	event := &nostr.Event{
		PubKey:    conn.PubKey,
		Kind:      24133,
		CreatedAt: time.Now(),
		Tags:      nostr.Tags{{"p", conn.AppID}},
		Content:   content,
	}
	err = event.Sign(conn.KeyInfo.Privkey)

	relay.Publish(ctx, *event)

	return nil
}
