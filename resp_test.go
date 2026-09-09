package main

import (
	"strings"
	"testing"
)

func TestReadline(t *testing.T) {
	line := "+OK\r\n"
	want := "+OK"

	resp := NewResp(strings.NewReader(line))
	got, _, err := resp.readLine()

	AssertNoError(t, err)
	AssertResult(t, string(got), want)
}

func TestReadInteger(t *testing.T) {
	line := "3\r\n"
	want := 3

	got, _, err  := NewResp(strings.NewReader(line)).readInteger()

	AssertNoError(t, err)
	AssertResult(t, got, want)
}

func TestReadBulk(t *testing.T) {
	line := "3\r\nset\r\n"
	expected := Value{typ: "bulk", bulk: "set"}

	got, err := NewResp(strings.NewReader(line)).readBulk()	

	AssertNoError(t, err)
	AssertValueResult(t, got, expected)
}

func TestReadArray(t *testing.T) {
	line := "3\r\n$3\r\nset\r\n$4\r\nname\r\n$3\r\ndan\r\n"
	expected := Value{typ: "array", array: []Value{{typ: "bulk",bulk: "set"}, {typ: "bulk", bulk: "name"}, {typ: "bulk", bulk: "dan"}}}

	got, err := NewResp(strings.NewReader(line)).readArray()

	AssertNoError(t, err)
	AssertValueResult(t, got, expected)
}

func TestMarshal(t *testing.T) {
	tests := []struct{
		name string 
		val Value
		want string
		}{
			{
				name: "converts given Value struct representing a string to RESP string",
				val: Value{
					typ: "string",
					str: "OK",
				},
				want: "+OK\r\n",
			},
			{
				name: "error",
				val: Value{
					typ: "error",
					str: "ERR something went wrong",
				},
				want: "-ERR something went wrong\r\n",
			},
			{
				name: "converts given value struct to RESP bulk string",
				val: Value{
					typ: "bulk",
					bulk: "hello",
				},
				want: "$5\r\nhello\r\n",
			},
			{
				name: "converts given value struct to RESP array string",
				val: Value{
					typ: "array",
					array: []Value{
						{
							typ:  "bulk",
							bulk: "GET",
						},
						{
							typ:  "bulk",
							bulk: "name",
						},
					},
				},
				want: "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n",
			},
			{
				name: "null",
				val: Value{
					typ: "null",
				},
				want: "$-1\r\n",
			},
		}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.val.Marshal()
			AssertResult(t, string(got), tc.want)
		})
	}
}