package traefik_taxsi2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nzin/traefik-taxsi2/internal/com"
)

// Config the plugin configuration.
type Config struct {
	Taxsi2Addr string `json:"taxsi2Addr" yaml:"taxsi2Addr"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Taxsi2Addr: "",
	}
}

// Demo a Demo plugin.
type Taxsi2Client struct {
	next http.Handler
	name string
	conn com.TaxsiConn
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.Taxsi2Addr) == 0 {
		return nil, fmt.Errorf("taxsi2Addr cannot be empty")
	}

	conn := com.NewHttpConn(config.Taxsi2Addr)

	return &Taxsi2Client{
		next: next,
		name: name,
		conn: conn,
	}, nil
}

func (t *Taxsi2Client) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if !t.conn.SubmitRequest(req) {
		http.Error(rw, "Forbidden: Access is denied", http.StatusForbidden)
		return
	}

	t.next.ServeHTTP(rw, req)
}
