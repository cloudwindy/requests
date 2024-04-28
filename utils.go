package requests

import (
	"bytes"
	"encoding/json"
	httphdrcustom "github.com/cloudwindy/requests/internal/httpheader"
	"github.com/cloudwindy/requests/internal/query"
	"github.com/mozillazg/go-httpheader"
	"net/http"
	"net/url"
	"strings"
)

func init() {
	query.EnableOmitEmptyByDefault = true
}

func MarshalQuery(v any) (s string, err error) {
	q, err := query.Values(v)
	if err != nil {
		return
	}
	return encode(q), nil
}

func UnmarshalHeaders(data http.Header, v any) (err error) {
	return httpheader.Decode(data.Clone(), v)
}

func MarshalHeaders(v any) (data http.Header, err error) {
	return httphdrcustom.Header(v)
}

func MarshalJson(v any) (str string, err error) {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// URL 编码
func encode(values url.Values) string {
	q := values.Encode()
	return strings.ReplaceAll(q, "%2A", "*")
}
