package main

import (
	"strconv"
	"strings"
	"time"
)

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
		case SETSTR:
			db.SETsMu.Lock()
			delete(db.SETs, key)
			db.SETsMu.Unlock()
		case HSETSTR:
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

func (db *DB) getExpiryType(key string) string {
	db.SETsMu.RLock()
	_, ok := db.SETs[key]
	db.SETsMu.RUnlock()

	if ok {
		return SETSTR
	}

	return HSETSTR
}

func (db *DB) deleteKey(expType string, key string) {
	switch expType {
	case SETSTR:
		db.SETsMu.Lock()
		delete(db.SETs, key)
		db.SETsMu.Unlock()
	case HSETSTR:
		db.HSETsMu.Lock()
		delete(db.HSETs, key)
		db.HSETsMu.Unlock()
	}
}