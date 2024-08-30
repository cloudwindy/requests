package requests

import (
	"context"
	"crypto/tls"
	"github.com/quic-go/quic-go/http3"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type Conn interface {
	RoundTrip(r *http.Request) (resp *http.Response, err error)
}

type conn struct {
	transport http.RoundTripper
}

type ConnConfig struct {
	Proxy     string
	DisableH2 bool
	EnableH3  bool
}

func NewConn(config ...ConnConfig) Conn {
	if len(config) == 0 {
		return newH2Conn()
	}
	if config[0].Proxy != "" || config[0].DisableH2 {
		return newH1Conn(config[0].Proxy)
	}
	if config[0].EnableH3 {
		return newH3Conn()
	}
	return newH2Conn()
}

func newH1Conn(proxy string) Conn {
	transport := &http.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			c, err := net.Dial(network, addr)
			if err != nil {
				return nil, err
			}
			uconn := utls.UClient(c, nil, utls.HelloRandomizedNoALPN)
			colonPos := strings.LastIndex(addr, ":")
			if colonPos == -1 {
				colonPos = len(addr)
			}
			uconn.SetSNI(addr[:colonPos])
			err = uconn.Handshake()
			return uconn, err
		},
		MaxIdleConnsPerHost: 128,
		MaxIdleConns:        128,
		MaxConnsPerHost:     128,
		ForceAttemptHTTP2:   false,
		TLSNextProto:        make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}
	if proxy != "" {
		u, err := url.Parse(proxy)
		if err != nil {
			panic(err)
		}
		transport.Proxy = http.ProxyURL(u)
	}
	return &conn{
		transport: transport,
	}
}

func newH2Conn() Conn {
	transport := &http2.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
			c, err := net.Dial(network, addr)
			if err != nil {
				return nil, err
			}
			uconn := utls.UClient(c, &utls.Config{NextProtos: cfg.NextProtos}, utls.HelloRandomized)
			colonPos := strings.LastIndex(addr, ":")
			if colonPos == -1 {
				colonPos = len(addr)
			}
			uconn.SetSNI(addr[:colonPos])
			err = uconn.Handshake()
			return uconn, err
		},
		MaxHeaderListSize:          0xffffffff,
		MaxReadFrameSize:           0x01000000,
		MaxDecoderHeaderTableSize:  0x00001000,
		MaxEncoderHeaderTableSize:  0x00001000,
		StrictMaxConcurrentStreams: true,
	}

	return &conn{
		transport: transport,
	}
}

func newH3Conn() Conn {
	return &conn{
		transport: &http3.RoundTripper{},
	}
}

func (c *conn) RoundTrip(r *http.Request) (resp *http.Response, err error) {
	return c.transport.RoundTrip(r)
}
