package com

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncoder(t *testing.T) {
	t.Run("simple eof", func(t *testing.T) {
		var buf bytes.Buffer
		e := NewEncoder(&buf)
		e.Eof()

		r := NewDecoder(bufio.NewReader(&buf))
		elt, err := r.ReadNextElement()
		assert.Nil(t, err)
		assert.Equal(t, byte(PROTO_EOF), elt.ElementType)

		_, err = r.ReadNextElement()
		assert.NotNil(t, err)
		assert.EqualError(t, err, "EOF")
	})
	t.Run("marshall/unmarshall", func(t *testing.T) {
		var buf bytes.Buffer
		e := NewEncoder(&buf)
		e.AddString(PROTO_METHOD, "GET")
		e.AddBody([]byte("foobar"))
		e.Eof()

		r := NewDecoder(bufio.NewReader(&buf))
		elt, err := r.ReadNextElement()
		assert.Nil(t, err)
		assert.Equal(t, byte(PROTO_METHOD), elt.ElementType)
		assert.Equal(t, "GET", string(elt.Payload))

		elt, err = r.ReadNextElement()
		assert.Nil(t, err)
		assert.Equal(t, byte(PROTO_BODY), elt.ElementType)
		assert.Equal(t, "foobar", string(elt.Payload))

		elt, err = r.ReadNextElement()
		assert.Nil(t, err)
		assert.Equal(t, byte(PROTO_EOF), elt.ElementType)
	})
}
