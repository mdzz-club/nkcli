package internal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

func UpdateRelays(ctx context.Context, key *KeyInfo, boots []string, w *sync.WaitGroup) {
	defer w.Done()

	now := time.Now()
	filters := nostr.Filters{{
		Kinds:   []int{0, 10002},
		Authors: []string{key.Pubkey},
		Limit:   2,
		Until:   &now,
	}}

	ech := make(chan *nostr.Event, len(boots))
	wg := new(sync.WaitGroup)
	db := ctx.Value("db").(*DB)

	relays := make([]string, 0)

	if key.Relays != nil {
		for u, v := range key.Relays {
			if !v.Read {
				continue
			}

			relays = append(relays, u)
		}
	} else {
		relays = append(relays, boots...)
	}

	for _, url := range boots {
		wg.Add(1)
		go subRelay(ctx, url, filters, ech, wg)
	}

	gch, cancel := context.WithCancel(ctx)
	go func() {
		for {
			select {
			case e := <-ech:
				if e == nil {
					break
				}
				fmt.Printf("Pubkey: %v Kind: %v\n", e.PubKey, e.Kind)

				if e.Kind == 0 {
					db.SaveEvent(bucketMetadatas, e)
				} else if e.Kind == 10002 {
					db.SaveEvent(bucketRelays, e)
				}
			case <-gch.Done():
				return
			}
		}
	}()

	wg.Wait()
	cancel()
}

func subRelay(ctx context.Context, relay string, filter nostr.Filters, ch chan<- *nostr.Event, wg *sync.WaitGroup) {
	defer wg.Done()

	conn, err := nostr.RelayConnect(context.Background(), relay)

	if err != nil {
		return
	}

	defer conn.Close()

	sub := conn.Subscribe(ctx, filter)
	defer sub.Unsub()

	for {
		select {
		case ev := <-sub.Events:
			ch <- ev
		case <-sub.EndOfStoredEvents:
			return
		case <-time.After(time.Second * 3):
			fmt.Printf("Relay %v connect timeout.\n", relay)
			return
		case <-ctx.Done():
			return
		}
	}
}
