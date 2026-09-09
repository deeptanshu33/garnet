package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestAOF(t *testing.T) {
	t.Run("test aof write", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "db.aof")

		aof, err := NewAof(path)

		if err != nil {
			t.Fatal(err)
		}
		defer aof.Close()

		value := Value{
			typ: "array",
			array: []Value{
				{
					typ:  "bulk",
					bulk: "SET",
				},
				{
					typ:  "bulk",
					bulk: "name",
				},
				{
					typ:  "bulk",
					bulk: "dan",
				},
			},
		}

		err = aof.Write(value)
		if err != nil {
			t.Fatal(err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		want := value.Marshal()

		if !bytes.Equal(data, want) {
			t.Errorf("expected %q, got %q", want, data)
		}
	})

	t.Run("test aof read", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "db.aof")

		aof, err := NewAof(path)

		if err != nil {
			t.Fatal(err)
		}
		defer aof.Close()

		file_data := "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n$5\r\nhello\r\n"

		_, err = aof.file.Write([]byte(file_data))
		if err != nil {
			t.Fatal(err)
		}

		_, err = aof.file.Seek(0, io.SeekStart)
		if err != nil {
			t.Fatal(err)
		}

		want := []Value{{
			typ: "array",
			array: []Value{
				{typ: "bulk", bulk: "GET"},
				{typ: "bulk", bulk: "name"},
			},
		}, {
			typ:  "bulk",
			bulk: "hello",
		}}

		var valueArr = []Value{}

		err = aof.Read(func(value Value) {
			valueArr = append(valueArr, value)
		})

		if err != nil {
			t.Error("did not expect an error but got one")
		}

		lenWant := len(want)
		lenGot := len(valueArr)

		if lenWant != lenGot {
			t.Errorf("expected equal length values, but got %d, %d", lenWant, lenGot)
		}

		length := len(want)
		for i := range length {
			AssertValueResult(t, want[i], valueArr[i])
		}
	})
}
