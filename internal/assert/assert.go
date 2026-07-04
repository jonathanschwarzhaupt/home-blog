package assert

import (
	"reflect"
	"strings"
	"testing"
)

func Equal[T any](t *testing.T, actual, expected T) {
	t.Helper()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}

func NotEqual[T any](t *testing.T, actual, expected T) {
	t.Helper()

	if reflect.DeepEqual(actual, expected) {
		t.Errorf("got: %v; want something different", actual)
	}
}

func True(t *testing.T, actual bool) {
	t.Helper()

	if !actual {
		t.Error("got: false; want: true")
	}
}

func False(t *testing.T, actual bool) {
	t.Helper()

	if actual {
		t.Error("got: true; want: false")
	}
}

func Nil(t *testing.T, actual any) {
	t.Helper()

	if !isNil(actual) {
		t.Errorf("got: %v; want: nil", actual)
	}
}

func NotNil(t *testing.T, actual any) {
	t.Helper()

	if isNil(actual) {
		t.Error("got: nil; want: not nil")
	}
}

func StringContains(t *testing.T, actual, expectedSubstring string) {
	t.Helper()

	if !strings.Contains(actual, expectedSubstring) {
		t.Errorf("got: %q; want it to contain: %q", actual, expectedSubstring)
	}
}

func isNil(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
