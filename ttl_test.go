package main

import (
	"strconv"
	"testing"
	"testing/synctest"
	"time"
)

func TestTTL(t *testing.T) {
	t.Run("set key expire returns ok", func(t *testing.T) {
		db := NewDB()
		args := []Value{
			{typ: "bulk", bulk: "name"}, 
			{typ: "bulk", bulk: "dan"}, 
			{typ: "bulk", bulk: "EX"},
			{typ: "bulk", bulk: "60"},
		}

		want := Value{typ: "string", str: "OK"}
		got := db.set(args)

		AssertValueResult(t, got, want)
	})

	t.Run("set key expire returns ok and key is deleted after specified time", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			db := NewDB()
			args := []Value{
				{typ: "bulk", bulk: "name"}, 
				{typ: "bulk", bulk: "dan"}, 
				{typ: "bulk", bulk: "PX"},
				{typ: "bulk", bulk: "10"},
			}

			db.set(args)
			time.Sleep(11*time.Millisecond)
			synctest.Wait()
			want := Value{typ: "null"}
			got := db.get([]Value{{typ: "bulk", bulk: "name"}})
			AssertValueResult(t, got, want)
		})
	})

	t.Run("hset key expire returns ok and key is deleted after specified time", func (t *testing.T)  {
		synctest.Test(t, func(t *testing.T) {
			db := NewDB()
			args := []Value{
				{typ: "bulk", bulk: "user"}, 
				{typ: "bulk", bulk: "name"}, 
				{typ: "bulk", bulk: "dan"},
				{typ: "bulk", bulk: "EX"},
				{typ: "bulk", bulk: "10"},
			}

			db.hset(args)
			time.Sleep(10*time.Second)
			synctest.Wait()
			want := Value{typ: "null"}
			got := db.hget([]Value{{typ: "bulk", bulk: "user"}, {typ: "bulk", bulk: "name"}})
			AssertValueResult(t, got, want)
		})
	})

	t.Run("correct value gets deleted after succesive set calls", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			db := NewDB()
			args := []Value{
				{typ: "bulk", bulk: "name"}, 
				{typ: "bulk", bulk: "dan"}, 
				{typ: "bulk", bulk: "EX"},
				{typ: "bulk", bulk: "10"},
			}

			db.set(args)

			time.Sleep(5 * time.Second)

			args = []Value{
				{typ: "bulk", bulk: "name"}, 
				{typ: "bulk", bulk: "manny"}, 
				{typ: "bulk", bulk: "EX"},
				{typ: "bulk", bulk: "60"},
			}

			db.set(args)

			time.Sleep(30*time.Second) 

			got := db.get([]Value{{typ: "bulk", bulk: "name"}})
			want := Value{typ: "bulk", bulk: "manny"}

			AssertValueResult(t, got, want)
		})
	})

	t.Run("correct value gets deleted after succesive hset calls", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			db := NewDB()
			args := []Value{
				{typ: "bulk", bulk: "user"}, 
				{typ: "bulk", bulk: "name"}, 
				{typ: "bulk", bulk: "dan"}, 
				{typ: "bulk", bulk: "EX"},
				{typ: "bulk", bulk: "10"},
			}

			db.hset(args)

			time.Sleep(5 * time.Second)

			args = []Value{
				{typ: "bulk", bulk: "user"}, 
				{typ: "bulk", bulk: "name"}, 
				{typ: "bulk", bulk: "manny"}, 
				{typ: "bulk", bulk: "EX"},
				{typ: "bulk", bulk: "60"},
			}
			db.hset(args)

			time.Sleep(30*time.Second) 

			got := db.hget([]Value{{typ: "bulk", bulk: "user"}, {typ: "bulk", bulk: "name"}})
			want := Value{typ: "bulk", bulk: "manny"}

			AssertValueResult(t, got, want)
		})
	})

	t.Run("new set without expiry, clears any running timer", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			db := NewDB()
			args := []Value{
				{typ: "bulk", bulk: "name"}, 
				{typ: "bulk", bulk: "dan"}, 
				{typ: "bulk", bulk: "EX"},
				{typ: "bulk", bulk: "10"},
			}

			db.set(args)
			time.Sleep(5*time.Second)

			args = []Value{
				{typ: "bulk", bulk: "name"}, 
				{typ: "bulk", bulk: "manny"},
			}
			db.set(args)

			time.Sleep(30*time.Second)

			got := db.get([]Value{{typ: "bulk", bulk: "name"}})
			want := Value{typ: "bulk", bulk: "manny"}

			AssertValueResult(t, got, want)
		})
	})

	t.Run("TTL command returns -2 if no entry is found for provided key", func(t *testing.T) {
		db := NewDB()
		valueArr := []Value{{typ: "bulk", bulk: "mykey"}}
		want := Value{typ: "integer", num: -2}
		got := db.ttl(valueArr)

		AssertValueResult(t, got, want)
	})

	t.Run("TTL command returns -1 if no expiry is found for provided key", func(t *testing.T) {
		db := NewDB()
		valArr := []Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "dan"}}
		db.set(valArr)

		want := Value{typ: "integer", num: -1}
		got := db.ttl([]Value{{typ: "bulk", bulk: "name"}})

		AssertValueResult(t, got, want)
	})

	t.Run("TTL command returns time left in seconds when expiry is found for provided key", func(t *testing.T) {
		synctest.Test(t, func (t *testing.T)  {
			db := NewDB()
			valArr := []Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "dan"}, {typ: "bulk", bulk: "ex"}, {typ: "bulk", bulk: "60"}}
			db.set(valArr)

			time.Sleep(35*time.Second)

			want := Value{typ: "integer", num: 25}
			got := db.ttl([]Value{valArr[0]})

			AssertValueResult(t, got, want)
		})
	})

	t.Run("PTTL command returns time left in milliseconds when expiry is found for provided key", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			db := NewDB()
			valArr := []Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "dan"}, {typ: "bulk", bulk: "px"}, {typ: "bulk", bulk: "500"}}
			db.set(valArr)

			time.Sleep(350*time.Millisecond)

			want := Value{typ: "integer", num: 150}
			got := db.pttl([]Value{{typ: "bulk", bulk: "name"}})

			AssertValueResult(t, got, want)
		})
	})

	t.Run("expire command returns 1 if key found", func(t *testing.T) {
		db := NewDB()
		valArr := []Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "dan"}}
		db.set(valArr)

		want := Value{typ: "integer", num: 1}
		got := db.expire([]Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "300"}})

		AssertValueResult(t, got, want)
	})

	t.Run("expire command returns 0 if key not found", func(t *testing.T) {
		db := NewDB()

		want := Value{typ: "integer", num: 0}
		got := db.expire([]Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "300"}})

		AssertValueResult(t, got, want)
	})

	t.Run("expire commands deletes key after said duration", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			db := NewDB()
			valArr := []Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "dan"}}
			db.set(valArr)

			want := Value{typ: "null"}
			db.expire([]Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "300"}})

			time.Sleep(302*time.Second)
			synctest.Wait()
			got := db.get([]Value{{typ: "bulk", bulk: "name"}})

			AssertValueResult(t, got, want)
		})
	})

	t.Run("expireat command return 0 if key not found", func(t *testing.T) {
		db := NewDB()
		got := db.expireat([]Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "1827456000"}})
		want := Value{typ: "integer", num: 0}
		AssertValueResult(t, got, want)
	})

	t.Run("expireat deletes a key after the specified duration", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			db := NewDB()
			db.set([]Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "dan"}})
			future_ts := time.Now().Unix() + 300
			db.expireat([]Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: strconv.FormatInt(future_ts, 10)}})

			time.Sleep(310*time.Second)

			got := db.get([]Value{{typ: "bulk", bulk: "name"}})
			want := Value{typ: "null"}

			AssertValueResult(t, got, want)
		})
	})

	t.Run("expireat deletes a key with a prior specified duration immediately", func(t *testing.T) {
		db := NewDB()
		db.set([]Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "dan"}})
		past_ts := time.Now().Unix() - 300
		db.expireat([]Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: strconv.FormatInt(past_ts, 10)}})
		got := db.get([]Value{{typ: "bulk", bulk: "name"}})
		want := Value{typ: "null"}
		AssertValueResult(t, got, want)
	})

	t.Run("persist removes ttl and key survives past original expiry", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			db := NewDB()
			db.set([]Value{
				{typ: "bulk", bulk: "name"},
				{typ: "bulk", bulk: "dan"},
				{typ: "bulk", bulk: "EX"},
				{typ: "bulk", bulk: "10"},
			})

			got := db.persist([]Value{{typ: "bulk", bulk: "name"}})
			AssertValueResult(t, got, Value{typ: "integer", num: 1})

			// TTL is gone immediately
			AssertValueResult(t, db.ttl([]Value{{typ: "bulk", bulk: "name"}}), Value{typ: "integer", num: -1})

			// ...and well past the original 10s expiry, the key is still there
			time.Sleep(15 * time.Second)
			synctest.Wait()

			got = db.get([]Value{{typ: "bulk", bulk: "name"}})
			AssertValueResult(t, got, Value{typ: "bulk", bulk: "dan"})
		})
	})

	t.Run("persist on missing key returns 0", func(t *testing.T) {
		db := NewDB()
		got := db.persist([]Value{{typ: "bulk", bulk: "ghost"}})
		AssertValueResult(t, got, Value{typ: "integer", num: 0})
	})

	t.Run("persist on key without ttl returns 0", func(t *testing.T) {
		db := NewDB()
		db.set([]Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "dan"}})

		got := db.persist([]Value{{typ: "bulk", bulk: "name"}})
		AssertValueResult(t, got, Value{typ: "integer", num: 0})
	})

	t.Run("persist wrong arity returns error", func(t *testing.T) {
		db := NewDB()
		got := db.persist([]Value{
			{typ: "bulk", bulk: "name"},
			{typ: "bulk", bulk: "extra"},
		})
		AssertValueResult(t, got, Value{typ: "error", str: "ERR wrong number of arguments for 'persist' command"})
	})

	t.Run("persist works on hashes", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			db := NewDB()
			db.hset([]Value{
				{typ: "bulk", bulk: "user"},
				{typ: "bulk", bulk: "name"},
				{typ: "bulk", bulk: "dan"},
				{typ: "bulk", bulk: "EX"},
				{typ: "bulk", bulk: "10"},
			})

			got := db.persist([]Value{{typ: "bulk", bulk: "user"}})
			AssertValueResult(t, got, Value{typ: "integer", num: 1})

			time.Sleep(15 * time.Second)
			synctest.Wait()

			got = db.hget([]Value{{typ: "bulk", bulk: "user"}, {typ: "bulk", bulk: "name"}})
			AssertValueResult(t, got, Value{typ: "bulk", bulk: "dan"})
		})
	})
}