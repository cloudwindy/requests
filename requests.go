package requests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Request struct {
	R         *http.Request
	transport http.RoundTripper
	sent      bool
}

func (req *Request) WithContext(ctx context.Context) *Request {
	req.R = req.R.WithContext(ctx)
	return req
}

func (req *Request) WithTransport(transport http.RoundTripper) *Request {
	req.transport = transport
	return req
}

func (req *Request) WithProxyTransport(us string) *Request {
	proxy, err := url.Parse(us)
	if err != nil {
		panic(err)
	}
	req.transport = &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}
	return req
}

func (req *Request) WithHost(host string) *Request {
	req.R.Host = host
	return req
}

func (req *Request) Send() (resp *http.Response, err error) {
	if req.sent {
		return nil, errors.New("already sent")
	}
	if req.transport == nil {
		req.transport = http.DefaultTransport
	}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout:   120 * time.Second,
		Transport: req.transport,
	}
	resp, err = client.Do(req.R)
	req.sent = true
	return
}

func (req *Request) WantBody() (body []byte, err error) {
	resp, err := req.Send()
	if err != nil {
		return
	}
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
		return
	}
	err = resp.Body.Close()
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = &Status{
			Code:     resp.StatusCode,
			Message:  resp.Status,
			Location: resp.Header.Get("Location"),
		}
		return
	}
	if resp.ContentLength == 0 {
		err = ErrEmptyBody
	}
	return
}

func (req *Request) WantJson(v any) error {
	body, err := req.WantBody()
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("%w: %s", err, string(body))
	}
	return nil
}

type Status struct {
	Code     int    `json:"status"`
	Message  string `json:"message"`
	Location string `json:"location"`
}

func (e *Status) Error() string {
	return fmt.Sprintf("unexpected response status: %s", e.Message)
}

var ErrEmptyBody = errors.New("empty body")
