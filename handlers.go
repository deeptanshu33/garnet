package main

import (
	"strconv"
	"strings"
	"sync"
	"time"
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

func (db *DB) clearTimer(key string) {
	db.ExpiryMu.Lock()
	db.TimerMu.Lock()

	if timer, ok := db.Timer[key]; ok {
		timer.Stop()
		delete(db.Timer, key)
	}

	delete(db.Expiry, key)

	db.TimerMu.Unlock()
	db.ExpiryMu.Unlock()
}

func validExpiry(value1, value2 Value) (int, bool) {
	if (value1.typ != "bulk" || value2.typ != "bulk") || (strings.ToUpper(value1.bulk) != "EX" && strings.ToUpper(value1.bulk) != "PX") {
		return -1, false
	}

	n, err := strconv.Atoi(value2.bulk)
	if err != nil {
		return -1, false
	}

	return n, n >= 0
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

func (db *DB) setExpiry(expType, key string, duration time.Duration) {
	expiryTime := time.Now().Add(duration)

	db.TimerMu.Lock()

	if oldTimer, ok := db.Timer[key]; ok {
		oldTimer.Stop()
	}

	expiryFunc := time.AfterFunc(duration, func() {
		db.ExpiryMu.Lock()
		currentExpiry, exists := db.Expiry[key]
		db.ExpiryMu.Unlock()

		if !exists || !currentExpiry.Equal(expiryTime) {
			return
		}

		switch expType {
		case "set":
			db.SETsMu.Lock()
			delete(db.SETs, key)
			db.SETsMu.Unlock()
		case "hset":
			db.HSETsMu.Lock()
			delete(db.HSETs, key)
			db.HSETsMu.Unlock()
		}

		db.ExpiryMu.Lock()
		delete(db.Expiry, key)
		db.ExpiryMu.Unlock()

		db.TimerMu.Lock()
		delete(db.Timer, key)
		db.TimerMu.Unlock()
	})

	db.Timer[key] = expiryFunc
	db.TimerMu.Unlock()

	db.ExpiryMu.Lock()
	db.Expiry[key] = expiryTime
	db.ExpiryMu.Unlock()
}

func (db *DB) ttl(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'ttl' command"}
	}

	if !db.keyExists(args[0].bulk) {
		return Value{typ: "string", str: "-2"}
	}

	if !db.expiryExists(args[0].bulk) {
		return Value{typ: "string", str: "-1"}
	}

	timeRemaining := db.getRemainingTime(args[0].bulk)
	timeRemainingSec := int64(timeRemaining.Seconds())

	return Value{typ: "string", str: strconv.Itoa(int(timeRemainingSec))}
}

func (db *DB) pttl(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'pttl' command"}
	}

	if !db.keyExists(args[0].bulk) {
		return Value{typ: "string", str: "-2"}
	}

	if !db.expiryExists(args[0].bulk) {
		return Value{typ: "string", str: "-1"}
	}

	timeRemaining := db.getRemainingTime(args[0].bulk)
	timeRemainingSec := int64(timeRemaining.Milliseconds())

	return Value{typ: "string", str: strconv.Itoa(int(timeRemainingSec))}
}

func (db *DB) getRemainingTime(key string) time.Duration {
	return time.Until(db.Expiry[key])
}

func (db *DB) keyExists(key string) bool {
	db.SETsMu.Lock()
	_, ok := db.SETs[key]
	db.SETsMu.Unlock()

	if ok {
		return true
	}

	db.HSETsMu.Lock()
	_, ok = db.HSETs[key]
	db.HSETsMu.Unlock()

	return ok
}

func (db *DB) expiryExists(key string) bool {
	db.ExpiryMu.RLock()
	_, ok := db.Expiry[key]
	db.ExpiryMu.RUnlock()

	return ok
}

func (db *DB) expire(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'expire' command"}
	}
	if !db.keyExists(args[0].bulk) {
		return Value{typ: "string", str: "0"}
	}

	return Value{typ: "string", str: "1"}
}

func NewHandlers(db *DB) map[string]HandlerFunc {
	return map[string]HandlerFunc{
		"PING": db.ping,
		"SET":  db.set,
		"GET":  db.get,
		"HSET": db.hset,
		"HGET": db.hget,
		"TTL":  db.ttl,
		"PTTL": db.pttl,
	}
}
