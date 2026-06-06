# goredis

A Redis clone written in Go from scratch using only the standard library.

## Run

```bash
git clone https://github.com/nomen06/goredis
cd goredis
go run .
```

Then in another terminal:

```bash
redis-cli
```

## Supported commands

| Command | Usage |
|---------|-------|
| PING | `PING` |
| SET | `SET key value` |
| GET | `GET key` |
| EXPIRE | `EXPIRE key seconds` |
| TTL | `TTL key` |
| HSET | `HSET hash key value` |
| HGET | `HGET hash key` |
| HGETALL | `HGETALL hash` |

## Structure

```
main.go       — TCP server on :6379, handles concurrent clients
resp.go       — RESP protocol parser and writer
handlers.go   — command logic and in-memory store
aof.go        — append-only file persistence, syncs to disk every second and replays on startup
```