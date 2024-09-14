package com

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

type HttpConn struct {
	endpoint                string
	failedCalls             int
	circuitBreakerOpenUntil time.Time
}

func NewHttpConn(endpoint string) *HttpConn {
	return &HttpConn{
		endpoint:                endpoint,
		failedCalls:             0,
		circuitBreakerOpenUntil: time.Now(),
	}
}

/*
SubmitRequest returns true if we can continue (WAF allows)
else false if we must forbid
*/
func (h *HttpConn) SubmitRequest(req *http.Request) bool {
	if h.failedCalls > 3 && time.Now().Before(h.circuitBreakerOpenUntil) {
		return true
	}
	// reset the circuit breaker
	if h.failedCalls > 3 {
		h.failedCalls = 0
		h.circuitBreakerOpenUntil = time.Now()
	}

	tc, err := NewTaxsiCom(req)
	if err != nil {
		// TBD log the error
		fmt.Println("error com:", err)
		return true
	}
	var buf bytes.Buffer
	err = tc.Marshall(&buf)
	if err != nil {
		// TBD log the error
		fmt.Println("error marshalling:", err)
		return true
	}

	resp, err := http.Post(h.endpoint, "application/octet-stream", &buf)
	if err != nil {
		fmt.Println("error submitting:", err)
		h.failedCalls++
		if h.failedCalls > 3 {
			h.circuitBreakerOpenUntil = time.Now().Add(30 * time.Second)
		}
		return true
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("passing")
		return true
	}
	if resp.StatusCode == http.StatusForbidden {
		fmt.Println("blocking")
		return false
	}

	// other status code
	h.failedCalls++
	fmt.Println("other status code:", resp.StatusCode)
	if h.failedCalls > 3 {
		h.circuitBreakerOpenUntil = time.Now().Add(30 * time.Second)
	}
	return true
}
