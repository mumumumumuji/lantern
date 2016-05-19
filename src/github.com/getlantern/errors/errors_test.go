package errors

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"

	"github.com/getlantern/context"
	"github.com/getlantern/hidden"
	"github.com/stretchr/testify/assert"
)

var (
	replaceNumbers = regexp.MustCompile("[0-9]+")
)

func TestFull(t *testing.T) {
	var firstErr *Error

	// Iterate past the size of the hidden buffer
	for i := 0; i < len(hiddenErrors)*2; i++ {
		ctx := context.Enter().Put("ca", 100)
		e := New("Hello %v", "There").Op("My Op").With("DaTa_1", i)
		ctx.Exit()
		if firstErr == nil {
			firstErr = e
		}
		assert.Equal(t, "Hello There", e.Error()[:11])
		ctx = context.Enter().Put("ca", 200).Put("cb", 200).Put("cc", 200)
		e3 := Wrap(fmt.Errorf("I'm wrapping your text: %v", e)).With("dATA+1", 3).With("cb", 300)
		ctx.Exit()
		assert.Equal(t, e, e3.cause, "Wrapping a regular error should have extracted the contained *Error")
		m := make(context.Map)
		e3.Fill(m)
		assert.Equal(t, i, m["data_1"], "Cause's data should dominate all")
		assert.Equal(t, 100, m["ca"], "Cause's context should dominate error")
		assert.Equal(t, 300, m["cb"], "Error's data should dominate its context")
		assert.Equal(t, 200, m["cc"], "Error's context should come through")
		assert.Equal(t, "My Op", e.data["error_op"], "Op should be available from cause")

		for _, call := range e3.callStack {
			t.Logf("at %v", call)
		}
	}

	e3 := Wrap(fmt.Errorf("I'm wrapping your text: %v", firstErr)).With("a", 2)
	assert.Nil(t, e3.cause, "Wrapping an *Error that's no longer buffered should have yielded no cause")
}

func TestNewWithCause(t *testing.T) {
	cause := New("World")
	outer := New("Hello %v", cause)
	assert.Equal(t, "Hello World", hidden.Clean(outer.Error()))
	assert.Equal(t, cause, outer.cause)

	// Make sure that stacktrace prints out okay
	buf := &bytes.Buffer{}
	print := outer.MultiLinePrinter()
	for {
		more := print(buf)
		buf.WriteByte('\n')
		if !more {
			break
		}
	}
	expected := `Hello World
  at github.com/getlantern/errors.TestNewWithCause:999
  at testing.tRunner:999
  at runtime.goexit:999
Caused by: World
  at github.com/getlantern/errors.TestNewWithCause:999
  at testing.tRunner:999
  at runtime.goexit:999
`
	assert.Equal(t, expected, replaceNumbers.ReplaceAllString(hidden.Clean(buf.String()), "999"))
}
