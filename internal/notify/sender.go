package notify

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/nicholas-fedor/shoutrrr/pkg/router"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
)

var (
	ErrInvalidURL  = errors.New("invalid notification service configuration")
	ErrSendFailed  = errors.New("notification provider rejected the request or could not be reached")
	ErrSendTimeout = errors.New("notification send deadline exceeded")
)

// ValidateURL validates through the installed Shoutrrr service implementation.
// Provider errors are intentionally replaced because they may include secrets.
func ValidateURL(raw string) error {
	var r router.ServiceRouter
	service, err := r.Locate(raw)
	if err != nil || !boundedSupport(service) {
		return ErrInvalidURL
	}
	return nil
}

func send(ctx context.Context, raw, message string) error {
	client := &boundedHTTPClient{ctx: ctx, client: &http.Client{Timeout: 30 * time.Second}}
	dial := func(_ context.Context, network, address string) (net.Conn, error) {
		conn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, network, address)
		if err != nil {
			return nil, err
		}
		bound := &boundedConn{Conn: conn, ctx: ctx}
		if deadline, ok := ctx.Deadline(); ok {
			_ = conn.SetDeadline(deadline)
		}
		bound.mu.Lock()
		bound.stop = context.AfterFunc(ctx, func() { _ = bound.Close() })
		bound.mu.Unlock()
		return bound, nil
	}
	r, err := router.NewWithOptions(nil, types.SenderOptions{HTTPClient: client, DialContext: dial})
	if err != nil {
		return ErrInvalidURL
	}
	service, err := r.Locate(raw)
	if err != nil {
		return ErrInvalidURL
	}
	if !boundedSupport(service) {
		return ErrInvalidURL
	}
	params := types.Params{}
	// Direct service invocation avoids the router's timeout goroutines, which
	// can outlive a send when a provider lacks ContextSender. Transport adapters
	// carry this request's context into every supported HTTP/TCP operation.
	if contextual, ok := service.(types.ContextSender); ok {
		err = contextual.SendContext(ctx, message, &params)
	} else {
		err = service.Send(message, &params)
	}
	if ctx.Err() != nil {
		return ErrSendTimeout
	}
	if err != nil {
		return ErrSendFailed
	}
	return nil
}

func boundedSupport(service types.Service) bool {
	// Shoutrrr's logger service has no external I/O and uses our discard logger.
	if service.GetID() == "logger" {
		return true
	}
	_, contextual := service.(types.ContextSender)
	_, httpBound := service.(types.HTTPClientSetter)
	_, tcpBound := service.(types.DialContextSetter)
	return contextual || httpBound || tcpBound
}

type boundedHTTPClient struct {
	ctx    context.Context
	client *http.Client
}

func (c *boundedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	response, err := c.client.Do(req.Clone(c.ctx))
	if err != nil {
		return nil, err
	}
	if response.Body != nil {
		response.Body = &limitedBody{Reader: io.LimitReader(response.Body, 1<<20), closer: response.Body}
	}
	return response, nil
}

type limitedBody struct {
	io.Reader
	closer io.Closer
}

func (b *limitedBody) Close() error { return b.closer.Close() }

type boundedConn struct {
	net.Conn
	ctx  context.Context
	stop func() bool
	once sync.Once
	mu   sync.Mutex
}

func (c *boundedConn) Close() error {
	var err error
	c.once.Do(func() {
		c.mu.Lock()
		if c.stop != nil {
			c.stop()
		}
		c.mu.Unlock()
		err = c.Conn.Close()
	})
	return err
}
func (c *boundedConn) limit(at time.Time) time.Time {
	if deadline, ok := c.ctx.Deadline(); ok && (at.IsZero() || deadline.Before(at)) {
		return deadline
	}
	return at
}
func (c *boundedConn) SetDeadline(at time.Time) error     { return c.Conn.SetDeadline(c.limit(at)) }
func (c *boundedConn) SetReadDeadline(at time.Time) error { return c.Conn.SetReadDeadline(c.limit(at)) }
func (c *boundedConn) SetWriteDeadline(at time.Time) error {
	return c.Conn.SetWriteDeadline(c.limit(at))
}
