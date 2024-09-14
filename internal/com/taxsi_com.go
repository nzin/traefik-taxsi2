package com

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type TaxsiCom struct {
	Headers    map[string][]string
	Body       []byte
	RemoteAddr string
	Method     string
	Url        *url.URL
}

func NewTaxsiCom(req *http.Request) (*TaxsiCom, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(body))

	t := &TaxsiCom{
		Headers:    make(map[string][]string),
		Body:       body,
		RemoteAddr: req.RemoteAddr,
		Method:     req.Method,
		Url:        req.URL,
	}
	for k, v := range req.Header {
		t.Headers[k] = v
	}
	return t, nil
}

func (t *TaxsiCom) Marshall(w io.Writer) error {
	e := NewEncoder(w)
	e.AddString(PROTO_METHOD, t.Method)
	for k, values := range t.Headers {
		for _, v := range values {
			e.AddString(PROTO_HEADER_KEY, k)
			e.AddString(PROTO_HEADER_VALUE, v)
		}
	}
	e.AddString(PROTO_REMOTEADDR, t.RemoteAddr)
	e.AddString(PROTO_URL, t.Url.String())
	e.AddBody(t.Body)

	e.Eof()
	return nil
}

func Unmarshall(r *bufio.Reader) (*TaxsiCom, error) {
	decoder := NewDecoder(r)
	body := []byte{}
	t := &TaxsiCom{
		Body:    body,
		Headers: make(map[string][]string),
	}
	for {
		elt, err := decoder.ReadNextElement()
		if err != nil {
			return nil, err
		}
		if elt.ElementType == PROTO_EOF {
			return t, nil
		}

		switch elt.ElementType {
		case PROTO_METHOD:
			t.Method = string(elt.Payload)
			continue

		case PROTO_URL:
			url, err := url.Parse(string(elt.Payload))
			if err != nil {
				return nil, err
			}
			t.Url = url
			continue

		case PROTO_HEADER_KEY:
			key := string(elt.Payload)
			v, err := decoder.ReadNextElement()
			if err != nil {
				return nil, err
			}
			if v.ElementType != PROTO_HEADER_VALUE {
				return nil, fmt.Errorf("a header value is expected after a header key, but got %v", v.ElementType)
			}
			if t.Headers[key] == nil {
				t.Headers[key] = []string{string(v.Payload)}
			} else {
				t.Headers[key] = append(t.Headers[key], string(v.Payload))
			}
			continue

		case PROTO_REMOTEADDR:
			t.RemoteAddr = string(elt.Payload)
			continue

		case PROTO_BODY:
			t.Body = elt.Payload
			continue

		default:
			// unknown protobyte
		}
	}
}
