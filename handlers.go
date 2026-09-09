package main

import (
	// "fmt"
	"sync"
)

type DB struct {
	SETs    map[string]string
	HSETs   map[string]map[string]string
	SETsMu  sync.RWMutex
	HSETsMu sync.RWMutex
}

func NewDB() *DB {
	return &DB{
		SETs: make(map[string]string),
		HSETs: make(map[string]map[string]string),
	}
}

func (db *DB) ping(args []Value) Value {
	return Value{typ: "string", str: "PONG"}
}

func (db *DB) set(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'set' command"}
	}

	key := args[0].bulk
	// fmt.Printf("key: %v\n", key)
	value := args[1].bulk
	// fmt.Printf("val: %v\n", value)
	db.SETsMu.Lock()
	db.SETs[key] = value
	db.SETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func (db *DB) hset(args []Value) Value {
	if len(args) != 3 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hset' command"}
	}

	mp := args[0].bulk
	key := args[1].bulk
	val := args[2].bulk

	if _, ok := db.HSETs[mp]; !ok {
		db.HSETsMu.Lock()
		db.HSETs[mp] = map[string]string{}
		db.HSETsMu.Unlock()
	}

	db.HSETsMu.Lock()
	db.HSETs[mp][key] = val
	db.HSETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func (db *DB) hget(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hget' command"}
	}

	mp := args[0].bulk
	key := args[1].bulk

	db.HSETsMu.RLock()
	val, ok := db.HSETs[mp][key]
	db.HSETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: val}
}

func (db *DB) get(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'get' command"}
	}

	key := args[0].bulk
	// fmt.Printf("key: %v\n", key)

	db.SETsMu.RLock()
	value, ok := db.SETs[key]
	db.SETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}
	// fmt.Printf("val: %v\n", value)
	return Value{typ: "bulk", bulk: value}
}

func NewHandlers(db *DB) map[string]func([]Value) Value {
	return map[string]func([]Value) Value{
		"PING": db.ping,
		"SET":  db.set,
		"GET":  db.get,
		"HSET": db.hset,
		"HGET": db.hget,
	}
}