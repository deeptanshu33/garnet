package main

import (
	// "fmt"
	"sync"
)

var SETs = map[string]string{}
var HSETs = map[string]map[string]string{}

var SETsMu = sync.RWMutex{}
var HSETsMu = sync.RWMutex{}

func ping(args []Value) Value {
	return Value{typ: "string", str: "PONG"}
}

func set(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'set' command"}
	}

	key := args[0].bulk
	// fmt.Printf("key: %v\n", key)
	value := args[1].bulk
	// fmt.Printf("val: %v\n", value)
	SETsMu.Lock()
	SETs[key] = value
	SETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func hset(args []Value) Value {
	if len(args) != 3 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hset' command"}
	}

	mp := args[0].bulk
	key := args[1].bulk
	val := args[2].bulk

	if _, ok := HSETs[mp]; !ok {
		HSETsMu.Lock()
		HSETs[mp] = map[string]string{}
		HSETsMu.Unlock()
	}

	HSETsMu.Lock()
	HSETs[mp][key] = val
	HSETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func hget(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hget' command"}
	}

	mp := args[0].bulk
	key := args[1].bulk

	HSETsMu.RLock()
	val, ok := HSETs[mp][key]
	HSETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: val}
}

func get(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'get' command"}
	}

	key := args[0].bulk
	// fmt.Printf("key: %v\n", key)
	
	SETsMu.RLock()
	value, ok := SETs[key]
	SETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}
	// fmt.Printf("val: %v\n", value)
	return Value{typ: "bulk", bulk: value}
}

var Handlers = map[string]func([]Value) Value {
	"PING": ping,
	"SET": set,
	"GET": get,
	"HSET": hset,
	"HGET": hget,
}