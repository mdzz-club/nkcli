package nkcli

import (
	"encoding/json"
	"errors"

	"github.com/nbd-wtf/go-nostr"
)

var (
	errInvalidKind = errors.New("Invalid event kind")
)

func parseEvent(buf []byte) (*nostr.Event, error) {
	var e *nostr.Event
	err := json.Unmarshal(buf, &e)

	if err != nil {
		return nil, err
	}

	return e, nil
}

func getUserMeta(e *nostr.Event) (*KeyMetadata, error) {
	meta := new(KeyMetadata)
	err := json.Unmarshal([]byte(e.Content), &meta)

	if err != nil {
		return nil, err
	}

	return meta, nil
}

func getRelayMap(e *nostr.Event) (RelayMap, error) {
	result := make(RelayMap)
	err := json.Unmarshal([]byte(e.Content), result)

	if err != nil {
		return nil, err
	}

	return result, nil
}