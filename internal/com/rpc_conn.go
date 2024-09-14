package com

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"time"
)

type RpcConn struct {
	connPool                *ConnPool
	failedCalls             int
	circuitBreakerOpenUntil time.Time
}

func NewRpcConn(rpcAddr string) (*RpcConn, error) {
	factory := func() (net.Conn, error) {
		return net.DialTimeout("tcp", rpcAddr, 5*time.Second)
	}

	// Create a new gRPC-like connection pool with a maximum of 5 connections
	connPool, err := NewConnPool(factory, 50)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %v", err)
	}

	return &RpcConn{
		connPool: connPool,
	}, nil
}

func (r *RpcConn) SubmitRequest(req *http.Request) bool {
	if r.failedCalls > 3 && time.Now().Before(r.circuitBreakerOpenUntil) {
		return true
	}
	// reset the circuit breaker
	if r.failedCalls > 3 {
		r.failedCalls = 0
		r.circuitBreakerOpenUntil = time.Now()
	}

	c, err := r.connPool.Get()
	if err != nil {
		return true
	}
	defer r.connPool.Put(c)

	tc, err := NewTaxsiCom(req)
	if err != nil {
		// TBD log the error
		return true
	}
	var buf bytes.Buffer
	err = tc.Marshall(&buf)
	if err != nil {
		// TBD log the error
		return true
	}

	_, err = c.Conn.Write(buf.Bytes())
	if err != nil {
		r.failedCalls++
		if r.failedCalls > 3 {
			r.circuitBreakerOpenUntil = time.Now().Add(30 * time.Second)
		}
		return true
	}
	response := []byte{0}
	n, err := c.Conn.Read(response)
	if n == 0 || err != nil {
		r.failedCalls++
		if r.failedCalls > 3 {
			r.circuitBreakerOpenUntil = time.Now().Add(30 * time.Second)
		}
		return true
	}
	// one byte: 0 == false, else true
	if response[0] == 0 {
		return false
	}
	return true
}
