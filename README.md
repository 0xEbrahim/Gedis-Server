# Gedis - A Redis-like In-Memory Database Server in Go

Gedis is a lightweight, Redis-compatible in-memory database server written in Go. It implements a subset of Redis commands and protocols, providing key-value storage, list operations, and hash data structures with persistence support.

## Features

###  Data Types Supported
- **Strings**: Basic key-value storage with expiration support
- **Lists**: Linked list operations (LPUSH, RPUSH, LPOP, RPOP, LRANGE, etc.)
- **Hashes**: Field-value maps within keys (HSET, HGET, HGETALL, etc.)

###  Core Features
- **RESP (Redis Serialization Protocol)** compliant
- **Persistence** through Append-Only File (AOF)
- **Thread-safe** operations with mutex locking
- **Key expiration** with TTL support
- **Atomic operations** on all data types
- **Graceful shutdown** with data flushing

###  Supported Commands

#### Key Operations
- `SET key value` - Set a string value
- `GET key` - Get a string value
- `KEYS` - List all keys
- `DEL key` - Delete a key
- `EXPIRE key seconds` - Set key expiration
- `TYPE key` - Get key type
- `RENAME old new` - Rename a key

#### List Operations
- `LPUSH key value [value...]` - Push to left
- `RPUSH key value [value...]` - Push to right
- `LPOP key` - Pop from left
- `RPOP key` - Pop from right
- `LLEN key` - Get list length
- `LINDEX key index` - Get element by index
- `LSET key index value` - Set element by index
- `LREM key count value` - Remove elements
- `LRANGE key start end` - Get range of elements

#### Hash Operations
- `HSET key field value [field value...]` - Set hash fields
- `HGET key field` - Get hash field
- `HGETALL key` - Get all fields and values
- `HKEYS key` - Get all field names
- `HVALS key` - Get all values
- `HLEN key` - Get number of fields
- `HEXISTS key field` - Check field existence
- `HDEL key field [field...]` - Delete fields

#### Server Operations
- `PING` - Server health check
- `ECHO message` - Echo back message
- `FLUSHALL` - Clear all data

## Installation

### Prerequisites
- Go 1.16 or higher
- Git

### Building from Source

```bash
# Clone the repository
git clone https://https://github.com/0xEbrahim/Gedis-Server
cd Gedis-Server

# Build the project
go build -o gedis ./

# Run the server (default port: 6379)
./gedis

# Run on specific port
./gedis -p <port>
```
