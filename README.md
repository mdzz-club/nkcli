# nkcli

Nostr key manager CLI tool.

## Features

- Multiple key management
- [NIP-46](https://github.com/nostr-protocol/nips/blob/master/46.md) support
- [NIP-06](https://github.com/nostr-protocol/nips/blob/master/06.md) support
- Encrypt your private key for security

## Install

Install via go install:

```
go install github.com/mdzz-club/nkcli@latest
```

## Usage

```
$ nkcli help
NAME:
   nkcli - Manage Nostr keys

USAGE:
   nkcli [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
   generate, g  Generate a new key
   list, l      List keys
   update, u    Update keys metadata and relay list
   import, i    Import your key
   remove       Remove your key and connected sessions
   connect, c   Create new connection via nostrconnect://
   disconnect   Disconnect and remove connection
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --db value, -d value  Database file (default: "/Users/boloto/.nkclidb") [$NKCLI_DB]
   --help, -h            show help
   --version, -v         print the version
```
