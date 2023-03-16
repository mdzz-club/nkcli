package nkcli

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/nbd-wtf/go-nostr"
)

type KeyInfo struct {
	Pubkey   string
	Privkey  string
	Metadata *KeyMetadata
	Relays   RelayMap
}

type KeyMetadata struct {
	Name     string `json:"display_name,omitempty"`
	Username string `json:"name,omitempty"`
	Nip05    string `json:"nip05,omitempty"`
}

type RelayAttr struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
}

type ConnMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Url         string `json:"url,omitempty"`
}

type Connection struct {
	AppID    string        `json:"appid"`
	Relay    string        `json:"relay"`
	PubKey   string        `json:"pubkey"`
	Metadata *ConnMetadata `json:"metadata"`
	Allows   []string      `json:"allows"`
	Acked    bool          `json:"acked"`
	KeyInfo  *KeyInfo      `json:"-"`
}

type RelayMap map[string]*RelayAttr

type DB struct {
	Db *bolt.DB
}

var (
	errDataNotFound      = errors.New("Data not found")
	errKeyNotFound       = errors.New("Pubkey not found")
	errInvalidPassphrase = errors.New("Passphrase is wrong")
)

var (
	bucketKeys        = []byte("keys")
	bucketRelays      = []byte("relays")
	bucketMetadatas   = []byte("metadatas")
	bucketConnections = []byte("connections")
)

func Open(p string) (*DB, error) {
	db, err := bolt.Open(p, 0600, nil)

	if err != nil {
		return nil, err
	}

	db.Update(func(tx *bolt.Tx) error {
		if _, err = tx.CreateBucketIfNotExists(bucketKeys); err != nil {
			return err
		}

		if _, err = tx.CreateBucketIfNotExists(bucketMetadatas); err != nil {
			return err
		}

		if _, err = tx.CreateBucketIfNotExists(bucketRelays); err != nil {
			return err
		}

		if _, err = tx.CreateBucketIfNotExists(bucketConnections); err != nil {
			return err
		}

		return nil
	})

	return &DB{db}, nil
}

func (d *DB) Close() error {
	return d.Db.Close()
}

func (d *DB) List() (keys []*KeyInfo, err error) {
	err = d.Db.View(func(t *bolt.Tx) error {
		c := t.Bucket(bucketKeys).Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			info := new(KeyInfo)
			info.Pubkey = hex.EncodeToString(k)

			buf := t.Bucket(bucketMetadatas).Get(k)
			if e, err := parseEvent(buf); err == nil {
				info.Metadata, _ = getUserMeta(e)
			}

			buf = t.Bucket(bucketRelays).Get(k)
			if e, err := parseEvent(buf); err == nil {
				info.Relays, _ = getRelayMap(e)
			}

			keys = append(keys, info)
		}

		return nil
	})

	return
}

func (d *DB) Has(key string) bool {
	return d.Db.View(func(tx *bolt.Tx) error {
		k, err := hex.DecodeString(key)

		if err != nil {
			return err
		}

		if tx.Bucket(bucketKeys).Get(k) == nil {
			return errors.New("not exists")
		}

		return nil
	}) == nil
}

func (d *DB) GetKey(pub string, pass []byte) (*KeyInfo, error) {
	result := new(KeyInfo)
	pubkey, err := hex.DecodeString(pub)

	if err != nil {
		return nil, err
	}

	err = d.Db.View(func(tx *bolt.Tx) error {
		priv := tx.Bucket(bucketKeys).Get(pubkey)

		if priv == nil {
			return errKeyNotFound
		}

		if pass != nil {
			rawPriv, err := Decrypt(priv, pass)

			if err != nil {
				return errInvalidPassphrase
			}

			result.Privkey = hex.EncodeToString(rawPriv)
		}

		result.Pubkey = hex.EncodeToString(pubkey)

		return nil
	})

	if err != nil {
		return nil, err
	}

	eb, _ := d.getDataById(bucketRelays, pubkey)

	if eb != nil {
		var e *nostr.Event

		if e, err = parseEvent(eb); err != nil {
			return nil, err
		}

		if res, err := getRelayMap(e); err == nil {
			result.Relays = res
		} else {
			return nil, err
		}
	}

	eb, _ = d.getDataById(bucketMetadatas, pubkey)

	if eb != nil {
		var e *nostr.Event

		if e, err = parseEvent(eb); err != nil {
			return nil, err
		}

		if meta, err := getUserMeta(e); err == nil {
			result.Metadata = meta
		} else {
			return nil, err
		}
	}

	return result, nil
}

func (d *DB) getDataById(bucket []byte, id []byte) (result []byte, err error) {
	err = d.Db.View(func(tx *bolt.Tx) error {
		if result = tx.Bucket(bucket).Get(id); result == nil {
			return errDataNotFound
		}

		return nil
	})

	return
}

func (d *DB) SaveKey(pub string, priv []byte) error {
	key, err := hex.DecodeString(pub)

	if err != nil {
		return err
	}

	return d.saveData(bucketKeys, key, priv)
}

func (d *DB) SaveRelays(key []byte, relays []string) error {
	return d.saveData(bucketRelays, key, []byte(strings.Join(relays, ",")))
}

func (d *DB) SaveMetadata(key []byte, metadata string) error {
	return d.saveData(bucketRelays, key, []byte(metadata))
}

func (d *DB) SaveEvent(bucket []byte, event *nostr.Event) error {
	buf, err := json.Marshal(event)

	if err != nil {
		return err
	}

	key, err := hex.DecodeString(event.PubKey)

	if err != nil {
		return err
	}

	return d.Db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucket).Put(key, buf)
	})
}

func (d *DB) saveData(bucket []byte, key []byte, data []byte) error {
	return d.Db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucket).Put(key, data)
	})
}

func (d *DB) Remove(key []byte) (err error) {
	err = d.Db.Update(func(tx *bolt.Tx) error {
		tx.Bucket(bucketKeys).Delete(key)

		tx.Bucket(bucketMetadatas).Delete(key)

		tx.Bucket(bucketRelays).Delete(key)

		b := tx.Bucket(bucketConnections)
		c := b.Cursor()
		pubkey := hex.EncodeToString(key)

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var j *Connection

			if err = json.Unmarshal(v, &j); err != nil {
				return err
			}

			if j.PubKey == pubkey {
				if err = b.Delete(k); err != nil {
					return err
				}
			}
		}

		return nil
	})
	return
}

func (d *DB) ListConnection() (list []*Connection, err error) {
	err = d.Db.View(func(t *bolt.Tx) error {
		c := t.Bucket(bucketConnections).Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			data := new(Connection)
			err = json.Unmarshal(v, data)

			if err != nil {
				return err
			}

			list = append(list, data)
		}

		return nil
	})

	return
}

func (d *DB) SetConnection(c *Connection) error {
	return d.Db.Update(func(tx *bolt.Tx) error {
		key, err := hex.DecodeString(c.AppID)
		if err != nil {
			return err
		}

		jsonStr, err := json.Marshal(c)
		if err != nil {
			return err
		}

		return tx.Bucket(bucketConnections).Put(key, jsonStr)
	})
}

func (d *DB) Disconnect(id string) error {
	key, err := hex.DecodeString(id)

	if err != nil {
		return err
	}

	return d.Db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketConnections).Delete(key)
	})
}
