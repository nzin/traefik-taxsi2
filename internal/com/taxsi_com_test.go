package com

import (
	"bufio"
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCom(t *testing.T) {
	t.Run("simple request marshal", func(t *testing.T) {
		method := "POST"
		url := "http://example.com/api"
		body := []byte(`{"key": "value"}`) // Example JSON body

		req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
		assert.Nil(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer some_token")

		var buf bytes.Buffer
		e, err := NewTaxsiCom(req)
		assert.Nil(t, err)

		err = e.Marshall(&buf)
		assert.Nil(t, err)

		d, err := Unmarshall(bufio.NewReader(&buf))
		assert.Nil(t, err)

		assert.Equal(t, "POST", d.Method)
		assert.Equal(t, 2, len(d.Headers))
		assert.Equal(t, "application/json", d.Headers["Content-Type"][0])
		assert.Equal(t, `{"key": "value"}`, string(d.Body))
	})
}
