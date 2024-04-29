package requests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Builder struct {
	method string
	scheme string
	host   string
	path   string
	query  string
	header http.Header
	body   io.Reader
}

func (rb *Builder) Url(us string) *Builder {
	u, err := url.Parse(us)
	if err != nil {
		panic(err)
	}
	rb.scheme = u.Scheme
	rb.host = u.Host
	rb.path = u.Path
	rb.query = u.RawQuery
	return rb
}

func (rb *Builder) Method(method string) *Builder {
	rb.method = method
	return rb
}

func (rb *Builder) Scheme(scheme string) *Builder {
	rb.scheme = scheme
	return rb
}

func (rb *Builder) Path(path string) *Builder {
	rb.path = path
	return rb
}

func (rb *Builder) Host(host string) *Builder {
	rb.host = host
	return rb
}

func (rb *Builder) Body(body []byte) *Builder {
	rb.body = bytes.NewReader(body)
	return rb
}

func (rb *Builder) BodyForm(object any) *Builder {
	body, err := MarshalQuery(object)
	if err != nil {
		panic(err)
	}
	rb.body = strings.NewReader(body)
	return rb
}

func (rb *Builder) BodyJson(object any) *Builder {
	body, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}
	rb.body = bytes.NewReader(body)
	return rb
}

func (rb *Builder) BodyReader(reader io.Reader) *Builder {
	rb.body = reader
	return rb
}

func (rb *Builder) Query(object any) *Builder {
	query, err := MarshalQuery(object)
	if err != nil {
		panic(err)
	}
	return rb.QueryString(query)
}

func (rb *Builder) QueryString(query string) *Builder {
	if rb.query != "" {
		rb.query += "&"
	}
	rb.query += query
	return rb
}

func (rb *Builder) Header(object any) *Builder {
	if rb.header == nil {
		rb.header = make(http.Header)
	}
	header, err := MarshalHeaders(object)
	if err != nil {
		panic(err)
	}
	for k, v := range header {
		rb.header[k] = v
	}
	return rb
}

func (rb *Builder) HeaderAdd(key string, value string) *Builder {
	if rb.header == nil {
		rb.header = make(http.Header)
	}
	rb.header.Add(key, value)
	return rb
}

func (rb *Builder) HeaderUpdate(header http.Header) *Builder {
	if rb.header == nil {
		rb.header = make(http.Header)
	}
	for k, v := range header {
		rb.header[k] = v
	}
	return rb
}

func (rb *Builder) BuildURL() (u *url.URL) {
	u = new(url.URL)
	u.Scheme = rb.scheme
	u.Host = rb.host
	u.Path = rb.path
	u.RawQuery = rb.query
	return
}

func (rb *Builder) Build() *Request {
	return rb.BuildWithContext(context.Background())
}

func (rb *Builder) BuildWithContext(ctx context.Context) *Request {
	req, err := http.NewRequestWithContext(ctx, rb.method, rb.BuildURL().String(), rb.body)
	if err != nil {
		panic(err)
	}
	req.Header = rb.header
	return &Request{R: req}
}
