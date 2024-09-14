package com

import "net/http"

type TaxsiConn interface {
	/*
	   SubmitRequest returns true if we can continue (WAF allows)
	   else if we must stop
	*/
	SubmitRequest(req *http.Request) bool
}
