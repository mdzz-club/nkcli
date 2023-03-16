package nkcli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"golang.org/x/crypto/ssh/terminal"
)

type ConnectRequest struct {
	Method string `json:"method"`
	ID     string `json:"id"`
	Params []any  `json:"params"`
	sub    *nostr.Subscription
	secret []byte
	Conn   *Connection
	ctx    context.Context
}

var (
	errUserRejected = errors.New("User rejected")
)

func (cr *ConnectRequest) CheckAllow(name string) error {
	if contains(cr.Conn.Allows, name) {
		return nil
	}

j1:
	fmt.Printf("\n🔑 Grant access to %v? [y(es)/n(o)/a(lways)]: ", name)

	switch Scanline() {
	case "y":
		return nil
	case "n":
		return errUserRejected
	case "a":
		cr.Conn.Allows = append(cr.Conn.Allows, name)

		db := cr.ctx.Value("db").(*DB)
		db.SetConnection(cr.Conn)

		return nil
	default:
		goto j1
	}
}

func (cr *ConnectRequest) Response(data any) error {
	var res map[string]any

	switch data.(type) {
	case error:
		res = map[string]any{"id": cr.ID, "error": data.(error).Error()}
	default:
		res = map[string]any{"id": cr.ID, "result": data}
	}

	jsonbuf, err := json.Marshal(res)

	if err != nil {
		return err
	}

	content, err := nip04.Encrypt(string(jsonbuf), cr.secret)

	if err != nil {
		return err
	}

	event := &nostr.Event{
		PubKey:    cr.Conn.PubKey,
		CreatedAt: time.Now(),
		Kind:      24133,
		Tags:      nostr.Tags{{"p", cr.Conn.AppID}},
		Content:   content,
	}
	err = event.Sign(cr.Conn.KeyInfo.Privkey)

	if err != nil {
		return err
	}

	cr.sub.Relay.Publish(cr.ctx, *event)

	return nil
}

func Serve(ctx context.Context, conn *Connection, ch chan<- *ConnectRequest, wg *sync.WaitGroup) {
	defer wg.Done()

	db := ctx.Value("db").(*DB)

	relay, err := nostr.RelayConnect(ctx, conn.Relay)
	defer relay.Close()

	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	sub := relay.Subscribe(ctx, nostr.Filters{{
		Kinds:   []int{24133},
		Authors: []string{conn.AppID},
		Tags:    nostr.TagMap{"p": []string{conn.PubKey}},
	}})
	defer sub.Unsub()

	if !conn.Acked {
		fmt.Print("Enter your passphrase to unlock your private key:")
		pass, err := terminal.ReadPassword(0)

		if err != nil {
			return
		}

		info, err := db.GetKey(conn.PubKey, pass)

		if err != nil {
			fmt.Printf("\nGet key info fail: %v\n", err)
			return
		}

		conn.KeyInfo = info

		rb := make([]byte, 16)
		rand.Read(rb)
		randomId := hex.EncodeToString(rb)
		data := map[string]any{"id": randomId, "method": "connect", "params": []string{conn.PubKey}}

		shared, err := nip04.ComputeSharedSecret(conn.AppID, conn.KeyInfo.Privkey)

		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}

		jstr, err := json.Marshal(data)

		if err != nil {
			return
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

		conn.Acked = true

		err = db.SetConnection(conn)

		if err != nil {
			fmt.Printf("Save connection status error: %v\n", err)
			return
		}

		fmt.Print("\nAck connection successful\n")
	}

j1:
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-sub.Events:
			if conn.KeyInfo == nil {
				fmt.Print("Enter your passphrase to unlock your private key:")
				pass, err := terminal.ReadPassword(0)

				if err != nil {
					return
				}

				conn.KeyInfo, err = db.GetKey(conn.PubKey, pass)

				if err != nil {
					return
				}
			}

			shared, err := nip04.ComputeSharedSecret(conn.AppID, conn.KeyInfo.Privkey)

			if err != nil {
				return
			}

			plain, err := nip04.Decrypt(e.Content, shared)

			if err != nil {
				return
			}

			var req *ConnectRequest
			err = json.Unmarshal([]byte(plain), &req)

			if err != nil {
				goto j1
			}

			if len(req.Method) == 0 {
				goto j1
			}

			req.sub = sub
			req.secret = shared
			req.Conn = conn
			req.ctx = ctx

			ch <- req
		}
	}
}

func contains(list []string, s string) bool {
	for _, i := range list {
		if i == s {
			return true
		}
	}

	return false
}
