package requests

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Request struct {
	R    *http.Request
	sent bool
}

func (req *Request) Send() (resp *http.Response, err error) {
	if req.sent {
		return nil, errors.New("already sent")
	}
	resp, err = http.DefaultClient.Do(req.R)
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
		err = errors.New("empty response body")
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
