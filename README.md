# goredis

A Redis-compatible server built from scratch in Go using only the standard library.

Implements the Redis Serialization Protocol (RESP) so any Redis client — including `redis-cli` — can connect to it out of the box.

## Features

- **RESP protocol** — full parser and writer, compatible with any Redis client
- **Core commands** — SET, GET with concurrent access via read-write locks
- **Hash commands** — HSET, HGET, HGETALL
- **Key expiry** — EXPIRE and TTL with expiry checking on access
- **Pub/Sub** — SUBSCRIBE, PUBLISH, UNSUBSCRIBE with per-client goroutines and Go channels for message fanout
- **AOF persistence** — appends every write to disk, replays on startup so data survives restarts
- **Concurrent clients** — each connection handled in its own goroutine
- **CLI client** — a custom Go client to connect and send commands without redis-cli

## Run the server

```bash
git clone https://github.com/nomen06/goredis
cd goredis
go run *.go
```

Connect with redis-cli:
```bash
redis-cli
127.0.0.1:6379> SET name john
OK
127.0.0.1:6379> GET name
"john"
```

## Run the CLI client

```bash
cd client
go run client.go
```

```
localhost:6379> SET foo bar
OK
localhost:6379> GET foo
bar
```

## Supported commands

| Command | Usage | Description |
|---------|-------|-------------|
| PING | `PING` | Health check |
| SET | `SET key value` | Store a string value |
| GET | `GET key` | Retrieve a value |
| EXPIRE | `EXPIRE key seconds` | Set expiry on a key |
| TTL | `TTL key` | Check remaining time on a key |
| HSET | `HSET hash key value` | Set a field in a hash |
| HGET | `HGET hash key` | Get a field from a hash |
| HGETALL | `HGETALL hash` | Get all fields from a hash |
| SUBSCRIBE | `SUBSCRIBE channel` | Subscribe to a channel |
| PUBLISH | `PUBLISH channel message` | Publish a message to a channel |
| UNSUBSCRIBE | `UNSUBSCRIBE channel` | Unsubscribe from a channel |

## Pub/Sub demo

Terminal 1:
```bash
redis-cli
127.0.0.1:6379> SUBSCRIBE news
```

Terminal 2:
```bash
redis-cli
127.0.0.1:6379> PUBLISH news "hello"
```

Terminal 1 instantly receives:
```
message
news
hello
```

## Project structure

```
main.go       — TCP server on :6379, concurrent client handling
resp.go       — RESP protocol parser and writer
handlers.go   — command logic, in-memory store, pub/sub
aof.go        — append-only persistence, syncs every second, replays on startup
client/
  client.go   — CLI client that speaks RESP to the server
```

## What I learned

- How the Redis wire protocol (RESP) actually works
- Building a concurrent TCP server in Go with goroutines
- Using `sync.RWMutex` to allow concurrent reads without blocking
- How pub/sub fan-out works using Go channels
- How databases survive restarts with append-only logs