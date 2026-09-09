package main

import "testing"

func TestSet(t *testing.T) {
	db := NewDB()
	args := []Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "dan"}}

	want := Value{typ: "string", str: "OK"}
	got := db.set(args)

	AssertValueResult(t, got, want)
}

func TestGet(t *testing.T) {
	db := NewDB()
	args := []Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "dan"}}
	db.set(args)
	
	args = []Value{{typ: "bulk", bulk: "name"}}
	want := Value{typ: "bulk", bulk: "dan"}
	got := db.get(args)

	AssertValueResult(t, got, want)
}

func TestHset(t *testing.T) {
	db := NewDB()
	args := []Value{{typ: "bulk", bulk: "directory"}, {typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "dan"}}

	got := db.hset(args)
	want := Value{typ: "string", str: "OK"}

	AssertValueResult(t, got, want)
}

func TestHget(t *testing.T) {
	db := NewDB()
	args := []Value{{typ: "bulk", bulk: "directory"}, {typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "dan"}}
	db.hset(args)

	args = []Value{{typ: "bulk", bulk: "directory"}, {typ: "bulk", bulk: "name"}}
	got := db.hget(args)
	want := Value{typ: "bulk", bulk: "dan"}

	AssertValueResult(t, got, want)
}

func TestSet_WrongNumberOfArgs(t *testing.T) {
	db := NewDB()

	// 0 args
	got := db.set([]Value{})
	want := Value{typ: "error", str: "ERR wrong number of arguments for 'set' command"}
	AssertValueResult(t, got, want)

	// 1 arg
	got = db.set([]Value{{typ: "bulk", bulk: "key"}})
	AssertValueResult(t, got, want)

	// 3 args
	got = db.set([]Value{
		{typ: "bulk", bulk: "key"},
		{typ: "bulk", bulk: "val"},
		{typ: "bulk", bulk: "extra"},
	})
	AssertValueResult(t, got, want)
}

// ─── GET error & null cases ───────────────────────────────────────────────

func TestGet_WrongNumberOfArgs(t *testing.T) {
	db := NewDB()

	// 0 args
	got := db.get([]Value{})
	want := Value{typ: "error", str: "ERR wrong number of arguments for 'get' command"}
	AssertValueResult(t, got, want)

	// 2 args
	got = db.get([]Value{
		{typ: "bulk", bulk: "a"},
		{typ: "bulk", bulk: "b"},
	})
	AssertValueResult(t, got, want)
}

func TestGet_KeyNotFound(t *testing.T) {
	db := NewDB()

	got := db.get([]Value{{typ: "bulk", bulk: "nonexistent"}})
	want := Value{typ: "null"}
	AssertValueResult(t, got, want)
}

// ─── HSET error cases ─────────────────────────────────────────────────────

func TestHset_WrongNumberOfArgs(t *testing.T) {
	db := NewDB()
	want := Value{typ: "error", str: "ERR wrong number of arguments for 'hset' command"}

	// 0 args
	got := db.hset([]Value{})
	AssertValueResult(t, got, want)

	// 2 args
	got = db.hset([]Value{
		{typ: "bulk", bulk: "hash"},
		{typ: "bulk", bulk: "key"},
	})
	AssertValueResult(t, got, want)

	// 4 args
	got = db.hset([]Value{
		{typ: "bulk", bulk: "hash"},
		{typ: "bulk", bulk: "key"},
		{typ: "bulk", bulk: "val"},
		{typ: "bulk", bulk: "extra"},
	})
	AssertValueResult(t, got, want)
}

// ─── HGET error & null cases ──────────────────────────────────────────────

func TestHget_WrongNumberOfArgs(t *testing.T) {
	db := NewDB()
	want := Value{typ: "error", str: "ERR wrong number of arguments for 'hget' command"}

	// 0 args
	got := db.hget([]Value{})
	AssertValueResult(t, got, want)

	// 1 arg
	got = db.hget([]Value{{typ: "bulk", bulk: "hash"}})
	AssertValueResult(t, got, want)

	// 3 args
	got = db.hget([]Value{
		{typ: "bulk", bulk: "hash"},
		{typ: "bulk", bulk: "key"},
		{typ: "bulk", bulk: "extra"},
	})
	AssertValueResult(t, got, want)
}

func TestHget_HashOrKeyNotFound(t *testing.T) {
	db := NewDB()

	// hash doesn't exist at all
	got := db.hget([]Value{
		{typ: "bulk", bulk: "nosuchhash"},
		{typ: "bulk", bulk: "nosuchkey"},
	})
	want := Value{typ: "null"}
	AssertValueResult(t, got, want)

	// hash exists but key doesn't
	db.hset([]Value{
		{typ: "bulk", bulk: "existinghash"},
		{typ: "bulk", bulk: "existingkey"},
		{typ: "bulk", bulk: "val"},
	})
	got = db.hget([]Value{
		{typ: "bulk", bulk: "existinghash"},
		{typ: "bulk", bulk: "nosuchkey"},
	})
	AssertValueResult(t, got, want)
}

// ─── PING (no args needed, but verify it works) ───────────────────────────

func TestPing(t *testing.T) {
	db := NewDB()
	got := db.ping([]Value{})
	want := Value{typ: "string", str: "PONG"}
	AssertValueResult(t, got, want)
}