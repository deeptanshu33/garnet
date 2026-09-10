# Garnet

A lightweight, multithreaded, in-memory key-value store built in Go. Garnet speaks the [RESP (Redis Serialization Protocol)](https://redis.io/docs/latest/develop/reference/protocol-spec/), so any standard Redis client — including `redis-cli` — works out of the box.

> **Why "Garnet"?** Like the gemstone, it's small, hard, and fast.

---

## Highlights

| | |
|---|---|
| **Protocol** | RESP2 — drop-in compatible with Redis clients |
| **Concurrency** | Goroutines + fine-grained `sync.RWMutex` per data structure |
| **Persistence** | Append-Only File (AOF) with background `fsync` every second |
| **Data Structures** | Strings (`SET`/`GET`) and Hashes (`HSET`/`HGET`) |
| **TTL Support** | Per-key expiry in seconds (`EX`) or milliseconds (`PX`) |
| **Zero Dependencies** | Pure Go standard library — no third-party packages |

---

## Getting Started

### Prerequisites

- [Go 1.26+](https://go.dev/dl/)

### Build & Run

```bash
# Clone the repository
git clone https://github.com/<your-username>/garnet.git
cd garnet

# Run directly
go run .

# Or build a binary
go build -o garnet .
./garnet
```

Garnet listens on **port 6379** by default — the same port as Redis.

### Connect with redis-cli

```bash
redis-cli

127.0.0.1:6379> PING
PONG

127.0.0.1:6379> SET greeting "hello world"
OK

127.0.0.1:6379> GET greeting
"hello world"
```

---

## Supported Commands

### Strings

| Command | Syntax | Description |
|---|---|---|
| **SET** | `SET key value [EX seconds \| PX milliseconds]` | Set a string value, optionally with an expiry |
| **GET** | `GET key` | Retrieve the value of a key (returns `nil` if not found) |

### Hashes

| Command | Syntax | Description |
|---|---|---|
| **HSET** | `HSET hash field value [EX seconds \| PX milliseconds]` | Set a field in a hash, optionally with an expiry on the hash |
| **HGET** | `HGET hash field` | Retrieve the value of a field within a hash |

### TTL & Expiry

| Command | Syntax | Description |
|---|---|---|
| **TTL** | `TTL key` | Time-to-live in **seconds**. Returns `-2` if the key doesn't exist, `-1` if no expiry is set |
| **PTTL** | `PTTL key` | Time-to-live in **milliseconds**. Same return semantics as `TTL` |

### Utility

| Command | Syntax | Description |
|---|---|---|
| **PING** | `PING` | Returns `PONG` — useful for health checks and connectivity tests |

---

## Architecture

### Project Structure

```
garnet/
├── main.go          # Entry point — TCP listener, connection loop, AOF replay
├── handlers.go      # Command implementations and DB struct
├── resp.go          # RESP protocol parser and serializer
├── aof.go           # Append-Only File persistence layer
├── testing.go       # Shared test helpers
├── handlers_test.go # Unit tests for command handlers
├── resp_test.go     # Unit tests for RESP serialization/deserialization
├── ttl_test.go      # Unit tests for TTL, PTTL, and expiry behaviour
├── database.aof     # AOF persistence file (generated at runtime)
└── go.mod           # Go module definition
```

### How It Works

```
┌──────────────┐        RESP over TCP         ┌──────────────────┐
│  Redis CLI   │  ◄──────────────────────────► │     Garnet       │
│  or any      │        (port 6379)            │                  │
│  RESP client │                               │  ┌────────────┐  │
└──────────────┘                               │  │ RESP Parser│  │
                                               │  └─────┬──────┘  │
                                               │        │         │
                                               │  ┌─────▼──────┐  │
                                               │  │  Handlers  │  │
                                               │  └─────┬──────┘  │
                                               │        │         │
                                               │  ┌─────▼──────┐  │
                                               │  │  In-Memory  │  │
                                               │  │   DB (maps) │  │
                                               │  └─────┬──────┘  │
                                               │        │         │
                                               │  ┌─────▼──────┐  │
                                               │  │    AOF      │  │
                                               │  │ (write log) │  │
                                               │  └────────────┘  │
                                               └──────────────────┘
```

1. **TCP Listener** — `main.go` opens a TCP socket on `:6379` and accepts connections.
2. **RESP Parser** — `resp.go` implements a streaming parser that reads RESP arrays and bulk strings from the wire, plus a serializer that marshals `Value` structs back into RESP format.
3. **Command Dispatch** — Incoming commands are uppercased and matched against a handler map built by `NewHandlers()`. Unknown commands return an empty string response.
4. **In-Memory Store** — The `DB` struct holds two primary maps: `SETs` (string → string) and `HSETs` (string → map[string]string), each guarded by its own `sync.RWMutex`.
5. **AOF Persistence** — Write commands (`SET`, `HSET`) are appended to `database.aof`. On startup, the AOF is replayed to restore state. A background goroutine calls `fsync` every second to flush writes to disk.

### Concurrency Model

Unlike Redis, which is single-threaded, Garnet is designed around Go's concurrency primitives:

- **Fine-grained locking** — Each logical data structure (`SETs`, `HSETs`, `Expiry`, `Timer`) has its own `sync.RWMutex`. Reads acquire shared locks (`RLock`); writes acquire exclusive locks (`Lock`). This means concurrent reads to different data structures never block each other.
- **Timer-based expiry** — `time.AfterFunc` spawns a goroutine per expiring key. When the timer fires, it re-validates the expiry timestamp before deleting — safely handling cases where a key has been overwritten or its TTL reset between scheduling and execution.
- **AOF sync** — A dedicated background goroutine periodically flushes the AOF file to disk, decoupled from the write path.

### RESP Protocol Implementation

Garnet implements the core RESP2 data types:

| Prefix | Type | Example |
|---|---|---|
| `+` | Simple String | `+OK\r\n` |
| `-` | Error | `-ERR unknown command\r\n` |
| `:` | Integer | `:1000\r\n` |
| `$` | Bulk String | `$5\r\nhello\r\n` |
| `*` | Array | `*2\r\n$3\r\nGET\r\n$4\r\nname\r\n` |

Null values are represented as `$-1\r\n`.

---

## Persistence

Garnet uses an **Append-Only File** strategy for durability:

- Every `SET` and `HSET` command is serialized in RESP format and appended to `database.aof`.
- On startup, the AOF is replayed command-by-command to rebuild the in-memory state.
- A background goroutine calls `file.Sync()` every second to ensure data is flushed to the OS.
- The AOF file is protected by a mutex, making concurrent writes safe.

> **Note:** AOF does not currently support compaction or rewriting. The file will grow unbounded with write-heavy workloads.

---

## Testing

Garnet has comprehensive unit tests covering command handlers, RESP serialization, and TTL behaviour — including deterministic time-based tests using Go's `testing/synctest` package.

```bash
# Run all tests
go test -v ./...

# Run a specific test suite
go test -v -run TestTTL
go test -v -run TestMarshal
```

### Test Coverage

| Area | File | What's Tested |
|---|---|---|
| Commands | `handlers_test.go` | SET, GET, HSET, HGET, PING — happy paths and error cases |
| RESP | `resp_test.go` | Line reading, integer parsing, bulk/array deserialization, marshalling |
| TTL | `ttl_test.go` | Expiry in EX/PX modes, timer replacement on overwrite, timer clearing, TTL/PTTL queries |
| AOF | `aof_test.go` | Persistence write/read round-trips |

---

## Differences from Redis

| | Redis | Garnet |
|---|---|---|
| **Threading** | Single-threaded event loop | Multithreaded (goroutines + mutexes) |
| **Language** | C | Go |
| **Data Structures** | Strings, Lists, Sets, Sorted Sets, Hashes, Streams, etc. | Strings, Hashes |
| **Persistence** | AOF + RDB snapshots | AOF only |
| **Cluster Support** | Yes | No |
| **Command Coverage** | 450+ commands | 7 commands |
| **Dependencies** | Multiple | Zero (stdlib only) |

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit your changes (`git commit -m 'Add my feature'`)
4. Push to the branch (`git push origin feature/my-feature`)
5. Open a Pull Request

When adding new commands:
1. Implement the handler method on the `DB` struct in `handlers.go`
2. Register it in the `NewHandlers()` map
3. If the command mutates state, add an AOF write in the main loop (`main.go`)
4. Add unit tests in the appropriate `*_test.go` file

---

## License

This project is open source. See the [LICENSE](LICENSE) file for details.
