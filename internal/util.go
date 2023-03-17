package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

type AppMeta struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
}

type NostrConnectInfo struct {
	Relay    string
	Metadata *AppMeta
	Pubkey   string
	Allows   map[string]bool
}

var (
	hexKeyRegexp         = regexp.MustCompile(`^[0-9a-f]{64}$`)
	errInvalidScheme     = errors.New("Invalid scheme")
	errInvalidPubkey     = errors.New("Invalid pubkey")
	errInvalidRelay      = errors.New("Invalid relay")
	errInvalidMetadata   = errors.New("Invalid metadata")
	errInvalidEventField = errors.New("Invalid event field")
)

func SerializeKeys(l []string) []string {
	result := make([]string, 0)

	for _, it := range l {
		if strings.HasPrefix(it, "nsec1") || strings.HasPrefix(it, "npub1") {
			_, res, err := nip19.Decode(it)

			if err != nil {
				continue
			}

			result = append(result, res.(string))
		} else {
			if regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(it) {
				continue
			}

			result = append(result, it)
		}
	}

	return result
}

func ParseURL(u string) (*NostrConnectInfo, error) {
	var metastr string
	info := new(NostrConnectInfo)
	obj, err := url.Parse(u)

	if err != nil {
		return nil, err
	}

	if obj.Scheme != "nostrconnect" {
		return nil, errInvalidScheme
	}

	if info.Pubkey = obj.Host; !hexKeyRegexp.Match([]byte(info.Pubkey)) {
		return nil, errInvalidPubkey
	}

	if info.Relay = obj.Query().Get("relay"); len(info.Relay) == 0 {
		return nil, errInvalidRelay
	}

	if metastr = obj.Query().Get("metadata"); len(metastr) == 0 {
		return nil, errInvalidMetadata
	}

	if err = json.Unmarshal([]byte(metastr), &info.Metadata); err != nil {
		return nil, err
	}

	return info, nil
}

func Scanline() string {
	var line string

	_, err := fmt.Scanln(&line)

	if err != nil {
		return ""
	}

	return line
}

func ParseEvent(obj map[string]any) (*nostr.Event, error) {
	e := new(nostr.Event)

	if v, ok := obj["pubkey"].(string); ok {
		e.PubKey = v
	}

	if v, ok := obj["id"].(string); ok {
		if len(e.PubKey) > 0 {
			return nil, errors.New("Invalid event ID")
		}

		e.ID = v
	}

	if v, ok := obj["kind"].(float64); !ok {
		return nil, errInvalidEventField
	} else {
		e.Kind = int(v)
	}

	if v, ok := obj["created_at"].(float64); !ok {
		return nil, errInvalidEventField
	} else {
		e.CreatedAt = time.Unix(int64(v), 0)
	}

	if v, ok := obj["content"].(string); !ok {
		return nil, errInvalidEventField
	} else {
		e.Content = v
	}

	e.Tags = parseTags(obj["tags"].([]interface{}))

	return e, nil
}

func parseTags(i []interface{}) (tags nostr.Tags) {
	for _, item := range i {
		tags = append(tags, nostr.Tag(item.([]string)))
	}
	return
}
