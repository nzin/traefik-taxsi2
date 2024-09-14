package com

import (
	"net"
	"sync"
	"time"
)

const (
	// https://medium.com/jamf-engineering/how-three-lines-of-configuration-solved-our-grpc-scaling-issues-in-kubernetes-ca1ff13f7f06
	MaxConnectionAge = 30 * time.Second
)

// ConnWrapper wraps net.Conn with additional metadata like creation time.
type ConnWrapper struct {
	Conn      net.Conn
	CreatedAt time.Time
}

// Pool manages a pool of net.Conn connections with channels.
type ConnPool struct {
	factory       func() (net.Conn, error) // Function to create new connections
	maxSize       int                      // Maximum pool size
	conns         chan *ConnWrapper        // Channel to store connections
	recreateMutex sync.Mutex               // Mutex to ensure recreation is thread-safe
}

/*
	 NewConnPool initializes a new connection pool with a channel. Usage:
		factory := func() (net.Conn, error) {
			return net.DialTimeout("tcp", config.Taxsi2Addr, 5*time.Second)
		}

		// Create a new gRPC-like connection pool with a maximum of 5 connections
		connPool, err := com.NewConnPool(factory, 50)
		if err != nil {
			return nil, fmt.Errorf("failed to create pool: %v", err)
		}
*/
func NewConnPool(factory func() (net.Conn, error), maxSize int) (*ConnPool, error) {
	p := &ConnPool{
		factory: factory,
		maxSize: maxSize,
		conns:   make(chan *ConnWrapper, maxSize),
	}

	// Pre-fill the pool with connections
	for i := 0; i < maxSize; i++ {
		connWrapper, err := p.createConnection()
		if err != nil {
			return nil, err
		}
		p.conns <- connWrapper
	}

	return p, nil
}

// createConnection creates a new ConnWrapper with a new net.Conn and tracks its creation time.
func (p *ConnPool) createConnection() (*ConnWrapper, error) {
	conn, err := p.factory()
	if err != nil {
		return nil, err
	}

	return &ConnWrapper{
		Conn:      conn,
		CreatedAt: time.Now(),
	}, nil
}

// Get retrieves a connection from the pool, recreating it if necessary.
func (p *ConnPool) Get() (*ConnWrapper, error) {
	connWrapper := <-p.conns

	// Check if the connection has exceeded MaxConnectionAge or is closed
	if time.Since(connWrapper.CreatedAt) > MaxConnectionAge || p.isClosed(connWrapper.Conn) {
		p.recreateMutex.Lock()
		defer p.recreateMutex.Unlock()

		// Close the old connection and create a new one
		connWrapper.Conn.Close()
		newConn, err := p.createConnection()
		if err != nil {
			return nil, err
		}
		connWrapper = newConn
	}

	return connWrapper, nil
}

// Put returns a connection to the pool
func (p *ConnPool) Put(connWrapper *ConnWrapper) {
	p.conns <- connWrapper
}

// isClosed checks if a connection is closed by attempting a non-blocking read
func (p *ConnPool) isClosed(conn net.Conn) bool {
	one := []byte{}
	if err := conn.SetReadDeadline(time.Now()); err != nil {
		return true
	}
	_, err := conn.Read(one)

	return err == net.ErrClosed
}

// Close closes all connections in the pool
func (p *ConnPool) Close() {
	close(p.conns)
	for connWrapper := range p.conns {
		connWrapper.Conn.Close()
	}
}
