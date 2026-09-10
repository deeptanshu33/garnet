package main

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	SETSTR  = "set"
	HSETSTR = "hset"
)

type DB struct {
	SETs  map[string]string
	HSETs map[string]map[string]string

	Expiry map[string]time.Time
	Timer  map[string]*time.Timer

	SETsMu   sync.RWMutex
	ExpiryMu sync.RWMutex
	TimerMu  sync.RWMutex
	HSETsMu  sync.RWMutex
}

type HandlerFunc func([]Value) Value

func NewDB() *DB {
	return &DB{
		SETs:   make(map[string]string),
		HSETs:  make(map[string]map[string]string),
		Expiry: make(map[string]time.Time),
		Timer:  make(map[string]*time.Timer),
	}
}

func (db *DB) ping(args []Value) Value {
	return Value{typ: "string", str: "PONG"}
}

func (db *DB) set(args []Value) Value {
	if len(args) < 2 || len(args) > 4 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'set' command"}
	}

	var exp int64 = -1
	if len(args) > 2 {
		if len(args) != 4 {
			return Value{typ: "error", str: "ERR wrong number of arguments for 'set' command"}
		}

		n, valid := validExpiry(args[2], args[3])
		exp = int64(n)
		if !valid {
			return Value{typ: "error", str: "ERR wrong arguments for 'expiry' command"}
		}

	}

	key := args[0].bulk

	value := args[1].bulk

	db.SETsMu.Lock()
	db.SETs[key] = value
	db.SETsMu.Unlock()

	if exp != -1 {
		switch strings.ToUpper(args[2].bulk) {
		case "EX":
			db.setExpiry("set", key, time.Duration(exp)*time.Second)

		case "PX":
			db.setExpiry("set", key, time.Duration(exp)*time.Millisecond)
		}
	} else {
		db.clearTimer(key)
	}

	return Value{typ: "string", str: "OK"}
}

func (db *DB) hset(args []Value) Value {
	if len(args) < 3 || len(args) > 5 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hset' command"}
	}

	var exp int64 = -1
	if len(args) > 3 {
		if len(args) != 5 {
			return Value{typ: "error", str: "ERR wrong number of arguments for 'hset' command"}
		}

		n, valid := validExpiry(args[3], args[4])
		exp = int64(n)
		if !valid {
			return Value{typ: "error", str: "ERR wrong arguments for 'expiry' command"}
		}
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

	if exp != -1 {
		switch strings.ToUpper(args[3].bulk) {
		case "EX":
			db.setExpiry("hset", mp, time.Duration(exp)*time.Second)

		case "PX":
			db.setExpiry("hset", mp, time.Duration(exp)*time.Millisecond)
		}
	} else {
		db.clearTimer(mp)
	}

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

func (db *DB) ttl(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'ttl' command"}
	}

	if !db.keyExists(args[0].bulk) {
		return Value{typ: "integer", num: -2}
	}

	if !db.expiryExists(args[0].bulk) {
		return Value{typ: "integer", num: -1}
	}

	timeRemaining := db.getRemainingTime(args[0].bulk)
	timeRemainingSec := int64(timeRemaining.Seconds())

	return Value{typ: "integer", num: int(timeRemainingSec)}
}

func (db *DB) pttl(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'pttl' command"}
	}

	if !db.keyExists(args[0].bulk) {
		return Value{typ: "integer", num: -2}
	}

	if !db.expiryExists(args[0].bulk) {
		return Value{typ: "integer", num: -1}
	}

	timeRemaining := db.getRemainingTime(args[0].bulk)
	timeRemainingSec := int64(timeRemaining.Milliseconds())

	return Value{typ: "integer", num: int(timeRemainingSec)}
}

func (db *DB) expire(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'expire' command"}
	}
	if !db.keyExists(args[0].bulk) {
		return Value{typ: "integer", num: 0}
	}

	db.clearTimer(args[0].bulk)
	expType := db.getExpiryType(args[0].bulk)

	d, err := strconv.Atoi(args[1].bulk)

	if err != nil {
		return Value{typ: "error", str: "ERR wrong arguments for 'expire' command"}
	}

	if d <= 0 {
		db.setExpiry(expType, args[0].bulk, time.Duration(0)*time.Second)
	} else {
		db.setExpiry(expType, args[0].bulk, time.Duration(d)*time.Second)
	}

	return Value{typ: "integer", num: 1}
}

func (db *DB) expireat(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'expireat' command"}
	}

	if !db.keyExists(args[0].bulk) {
		return Value{typ: "integer", num: 0}
	}

	db.clearTimer(args[0].bulk)

	ts, err := strconv.ParseInt(args[1].bulk, 10, 64)
	if err != nil {
		return Value{typ: "error", str: "ERR value is not an integer or out of range"}
	}

	duration := time.Until(time.Unix(ts, 0))

	expType := db.getExpiryType(args[0].bulk)

	if duration <= 0 {
		db.deleteKey(expType, args[0].bulk)
	} else {
		db.setExpiry(expType, args[0].bulk, duration)
	}

	return Value{typ: "integer", num: 1}
}

func (db *DB) persist(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'persist' command"}
	}

	if !db.keyExists(args[0].bulk) {
		return Value{typ: "integer", num: 0}
	}

	db.ExpiryMu.RLock()
	_, hasTTL := db.Expiry[args[0].bulk]
	db.ExpiryMu.RUnlock()

	if !hasTTL {
		return Value{typ: "integer", num: 0}
	}

	db.clearTimer(args[0].bulk)

	return Value{typ: "integer", num: 1}
}

func NewHandlers(db *DB) map[string]HandlerFunc {
	return map[string]HandlerFunc{
		"PING":     db.ping,
		"SET":      db.set,
		"GET":      db.get,
		"HSET":     db.hset,
		"HGET":     db.hget,
		"TTL":      db.ttl,
		"PTTL":     db.pttl,
		"EXPIRE":   db.expire,
		"EXPIREAT": db.expireat,
		"PERSIST":  db.persist,
	}
}
