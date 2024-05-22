package requests

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Request struct {
	R      *http.Request
	client *http.Client
	sent   bool
}

func (req *Request) WithProxy(us string) *Request {
	proxy, err := url.Parse(us)
	if err != nil {
		panic(err)
	}
	req.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
		},
	}
	return req
}

func (req *Request) Send() (resp *http.Response, err error) {
	if req.sent {
		return nil, errors.New("already sent")
	}
	if req.client == nil {
		req.client = http.DefaultClient
	}
	resp, err = req.client.Do(req.R)
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
