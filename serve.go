package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	nkcli "github.com/mdzz-club/nkcli/internal"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/nbd-wtf/go-nostr/nip26"
	"github.com/urfave/cli/v2"
)

func serveAction(c *cli.Context) error {
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
		fmt.Print(`You dont't have any connections.
Run 'nkcli generate' to generate a new keypair.
Run 'nkcli import [-raw] [nsec or mnemonic]' to import a key.
Run 'nkcli connect [url]' to create a connection.

Run 'nkcli help' get more.
`)
		return nil
	}

	fmt.Printf("Serving %v connections...\n", len(conns))

	ctx := context.WithValue(c.Context, "db", db)
	ctx, cancel := context.WithCancel(ctx)
	signalCh := make(chan os.Signal, 1)
	cMap := make(map[string]context.CancelFunc)

	go func() {
		<-signalCh
		cancel()
	}()

	reqCh := make(chan *nkcli.ConnectRequest, 10)
	wg := new(sync.WaitGroup)

	for _, item := range conns {
		wg.Add(1)
		c, f := context.WithCancel(ctx)
		cMap[item.AppID] = f
		go nkcli.Serve(c, item, reqCh, wg)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case req := <-reqCh:
				fmt.Printf("\n  🔔 Request ID: %v Method: %v\n", req.ID, req.Method)

				switch req.Method {
				case "describe":
					req.Response([]string{"describe", "get_public_key", "sign_event", "disconnect", "nip04_encrypt", "nip04_decrypt", "delegate"})
				case "get_public_key":
					if err = req.CheckAllow("get_public_key"); err != nil {
						req.Response(err)
						continue
					}

					req.Response(req.Conn.PubKey)
				case "sign_event":
					obj := req.Params[0].(map[string]interface{})
					tmpEv := new(nostr.Event)
					tmpEv.ID = obj["id"].(string)
					tmpEv.PubKey = obj["pubkey"].(string)
					tmpEv.Kind = int(obj["kind"].(float64))
					tmpEv.CreatedAt = time.Unix(int64(obj["created_at"].(float64)), 0)
					tmpEv.Content = obj["content"].(string)
					tmpEv.Tags = parseTags(obj["tags"].([]interface{}))

					fmt.Printf("Event detail:\n\n")

					printEvent(tmpEv)

					if err = req.CheckAllow("sign_event"); err != nil {
						req.Response(err)
						continue
					}

					if err = tmpEv.Sign(req.Conn.KeyInfo.Privkey); err != nil {
						req.Response(err)
						continue
					}

					req.Response(tmpEv.Sig)
				case "disconnect":
					fmt.Printf("%v Request disconnect.", req.Conn.AppID)

					if err = db.Disconnect(req.Conn.AppID); err != nil {
						continue
					}

					cancel := cMap[req.Conn.AppID]
					cancel()
				case "get_relays":
					if err = req.CheckAllow("get_relays"); err != nil {
						req.Response(err)
						continue
					}

					req.Response(req.Conn.KeyInfo.Relays)
				case "nip04_encrypt":
					pub, plain := req.Params[0].(string), req.Params[1].(string)

					fmt.Printf("Encrypt message to %v with following text:\n\n%v\n\n", pub, plain)

					if err = req.CheckAllow("nip04_encrypt"); err != nil {
						req.Response(err)
						continue
					}

					shared, err := nip04.ComputeSharedSecret(pub, req.Conn.KeyInfo.Privkey)

					if err != nil {
						req.Response(err)
						continue
					}

					ciphered, err := nip04.Encrypt(plain, shared)

					if err != nil {
						req.Response(err)
						continue
					}

					req.Response(ciphered)
				case "nip04_decrypt":
					pub, plain := req.Params[0].(string), req.Params[1].(string)

					fmt.Printf("Decrypt message from %v\n", pub)

					if err = req.CheckAllow("nip04_decrypt"); err != nil {
						req.Response(err)
						continue
					}

					shared, err := nip04.ComputeSharedSecret(pub, req.Conn.KeyInfo.Privkey)

					if err != nil {
						req.Response(err)
						continue
					}

					text, err := nip04.Decrypt(plain, shared)

					if err != nil {
						req.Response(err)
						continue
					}

					req.Response([]string{text})
				case "delegate":
					delegatee, conds := req.Params[0].(string), req.Params[1].(map[string]any)
					kind := int(conds["kind"].(float64))
					since := time.Unix(int64(conds["since"].(float64)), 0)
					until := time.Unix(int64(conds["until"].(float64)), 0)

					fmt.Printf("Delegate to %v with these conditions:\n\n", delegatee)
					fmt.Printf("  Kind: %v\n  Since: %v\n  Until: %v\n\n", kind, formatTime(&since), formatTime(&until))

					if err = req.CheckAllow("delegate"); err != nil {
						req.Response(err)
						continue
					}

					d, err := nip26.CreateToken(req.Conn.KeyInfo.Privkey, delegatee, []int{kind}, &since, &until)

					if err != nil {
						req.Response(err)
						continue
					}

					sig := d.Tag()[3]

					req.Response(map[string]string{
						"from": req.Conn.PubKey,
						"to":   delegatee,
						"cond": d.Conditions(),
						"sig":  sig,
					})
				}
			}
		}
	}()

	wg.Wait()

	return nil
}

func parseTags(i []interface{}) (tags nostr.Tags) {
	for _, item := range i {
		tags = append(tags, nostr.Tag(item.([]string)))
	}
	return
}

func printEvent(ev *nostr.Event) {
	str, _ := json.MarshalIndent(ev, "", "  ")
	fmt.Printf("%s\n", str)
}

func formatTime(t *time.Time) string {
	return fmt.Sprintf("%v (%v)", t.Unix(), t.Format(time.DateTime))
}
