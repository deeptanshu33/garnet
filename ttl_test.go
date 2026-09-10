package main

import (
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
		want := Value{typ: "string", str: "-2"}
		got := db.ttl(valueArr)

		AssertValueResult(t, got, want)
	})

	t.Run("TTL command returns -1 if no expiry is found for provided key", func(t *testing.T) {
		db := NewDB()
		valArr := []Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "dan"}}
		db.set(valArr)

		want := Value{typ: "string", str: "-1"}
		got := db.ttl([]Value{{typ: "bulk", bulk: "name"}})

		AssertValueResult(t, got, want)
	})

	t.Run("TTL command returns time left in seconds when expiry is found for provided key", func(t *testing.T) {
		synctest.Test(t, func (t *testing.T)  {
			db := NewDB()
			valArr := []Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "dan"}, {typ: "bulk", bulk: "ex"}, {typ: "bulk", bulk: "60"}}
			db.set(valArr)

			time.Sleep(35*time.Second)

			want := Value{typ: "string", str: "25"}
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

			want := Value{typ: "string", str: "150"}
			got := db.pttl([]Value{{typ: "bulk", bulk: "name"}})

			AssertValueResult(t, got, want)
		})
	})

	t.Run("expire command returns 1 if key found", func(t *testing.T) {
		db := NewDB()
		valArr := []Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "dan"}}
		db.set(valArr)

		want := Value{typ: "string", str: "1"}
		got := db.expire([]Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "300"}})

		AssertValueResult(t, got, want)
	})

	t.Run("expire command returns 0 if key not found", func(t *testing.T) {
		db := NewDB()

		want := Value{typ: "string", str: "0"}
		got := db.expire([]Value{{typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "300"}})

		AssertValueResult(t, got, want)
	})
}