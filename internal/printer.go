package internal

import (
	"fmt"
	"strings"

	"github.com/nbd-wtf/go-nostr/nip19"
)

func PrintKeyList(keys []*KeyInfo) {
	for i, k := range keys {
		printKey(i, k)
	}
}

func printKey(index int, key *KeyInfo) {
	npub, _ := nip19.EncodePublicKey(key.Pubkey)
	fmt.Printf("  %v. %v\n     %v\n     %v\n\n", index+1, keyName(key), npub, key.Pubkey)
}

func keyName(k *KeyInfo) string {
	if k.Metadata != nil {
		s := make([]string, 0)

		if len(k.Metadata.Name) > 0 {
			s = append(s, k.Metadata.Name)
		}

		if len(k.Metadata.Username) > 0 {
			s = append(s, "@"+k.Metadata.Username)
		}

		if len(k.Metadata.Nip05) > 0 {
			s = append(s, fmt.Sprintf("<%v>", k.Metadata.Nip05))
		}

		if len(s) == 0 {
			return "(no name)"
		}

		return strings.Join(s, " ")
	} else {
		return "(no name)"
	}
}
