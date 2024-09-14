package com

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	_ = iota
	PROTO_EOF
	PROTO_METHOD
	PROTO_URL
	PROTO_HEADER_KEY   // key
	PROTO_HEADER_VALUE // value
	PROTO_BODY
	PROTO_REMOTEADDR
)

type Element struct {
	ElementType byte
	Length      int32
	Payload     []byte
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		writer: w,
	}
}

type Encoder struct {
	writer io.Writer
}

func (e *Encoder) AddString(prototype byte, str string) {
	e.writer.Write([]byte{prototype})

	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(len(str)))
	e.writer.Write(b)

	e.writer.Write([]byte(str))
}

func (e *Encoder) AddBody(payload []byte) {
	e.writer.Write([]byte{PROTO_BODY})

	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(len(payload)))
	e.writer.Write(b)

	e.writer.Write(payload)
}

func (e *Encoder) Eof() {
	e.writer.Write([]byte{PROTO_EOF})
}

type Decoder struct {
	reader *bufio.Reader
}

func NewDecoder(r *bufio.Reader) *Decoder {
	return &Decoder{
		reader: r,
	}
}

func (d *Decoder) ReadNextElement() (*Element, error) {
	prototype, err := d.reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if prototype == PROTO_EOF {
		return &Element{
			ElementType: PROTO_EOF,
			Length:      0,
			Payload:     []byte{},
		}, nil
	}

	lengthByte := make([]byte, 4)
	s, err := io.ReadFull(d.reader, lengthByte)
	if err != nil {
		return nil, err
	}
	if s != 4 {
		return nil, fmt.Errorf("not able to read 4 bytes for the length")
	}
	length := binary.LittleEndian.Uint32(lengthByte)

	payload := make([]byte, length)
	s, err = io.ReadFull(d.reader, payload)
	if err != nil {
		return nil, err
	}
	if s != int(length) {
		return nil, fmt.Errorf("not able to read 4 bytes for the length")
	}
	return &Element{
		ElementType: prototype,
		Length:      int32(length),
		Payload:     payload,
	}, nil
}
