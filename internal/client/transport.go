package client

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// TransportOptions configures HTTP connection pooling and circuit breaking.
type TransportOptions struct {
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeoutSec  int
	CircuitThreshold    int
	CircuitCooldown     time.Duration
}

// pooledTransport wraps http.Transport with a simple circuit breaker.
type pooledTransport struct {
	base     *http.Transport
	mu       sync.Mutex
	failures int
	openUntil time.Time
	opts     TransportOptions
}

func newPooledTransport(opts TransportOptions) *pooledTransport {
	if opts.MaxIdleConns <= 0 {
		opts.MaxIdleConns = 100
	}
	if opts.MaxIdleConnsPerHost <= 0 {
		opts.MaxIdleConnsPerHost = 10
	}
	if opts.IdleConnTimeoutSec <= 0 {
		opts.IdleConnTimeoutSec = 90
	}
	if opts.CircuitThreshold <= 0 {
		opts.CircuitThreshold = 5
	}
	if opts.CircuitCooldown <= 0 {
		opts.CircuitCooldown = 30 * time.Second
	}
	base := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        opts.MaxIdleConns,
		MaxIdleConnsPerHost: opts.MaxIdleConnsPerHost,
		IdleConnTimeout:     time.Duration(opts.IdleConnTimeoutSec) * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return &pooledTransport{base: base, opts: opts}
}

func (t *pooledTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.beforeRequest(); err != nil {
		return nil, err
	}
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		t.recordFailure()
		return nil, err
	}
	if resp.StatusCode >= http.StatusInternalServerError {
		t.recordFailure()
	} else {
		t.recordSuccess()
	}
	return resp, err
}

func (t *pooledTransport) beforeRequest() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if time.Now().Before(t.openUntil) {
		return fmt.Errorf("API 熔断中，请 %s 后重试", t.openUntil.Format("15:04:05"))
	}
	return nil
}

func (t *pooledTransport) recordFailure() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.failures++
	if t.failures >= t.opts.CircuitThreshold {
		t.openUntil = time.Now().Add(t.opts.CircuitCooldown)
		t.failures = 0
	}
}

func (t *pooledTransport) recordSuccess() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.failures = 0
}
