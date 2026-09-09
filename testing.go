package main

import (
	"testing"
)

func AssertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("didn't expect an error but got %s", err.Error())
	}
}

func AssertResult(t testing.TB, got, want any) {
	t.Helper()
	if got!=want {
		t.Errorf("wanted %v, got %v", want, got)
	}
}

func AssertValueResult(t *testing.T, got, want Value) {
	t.Helper()
	if got.typ != want.typ {
		t.Errorf("expected type %q, got %q", want.typ, got.typ)
	}
	if got.str != want.str {
		t.Errorf("expected str %q, got %q", want.str, got.str)
	}
	if got.bulk != want.bulk {
		t.Errorf("expected bulk %q, got %q", want.bulk, got.bulk)
	}
}

