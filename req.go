package req

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Headers map[string]string
type Params map[string]interface{}
type Timeout time.Duration
type List []interface{}

var ErrorParamInvalid = errors.New("req param invalid")

type Req struct {
	method    string
	headers   Headers
	params    Params
	values    url.Values
	timeout   time.Duration
	tlsConfig *tls.Config
	url       string
	req       *http.Request
	client    *http.Client
}

func (r *Req) Get(url string, values ...interface{}) (*Resp, error) {
	r.method = "GET"
	r.url = url
	return r.Do(values...)
}

func (r *Req) GetJson(result interface{}, uri string, values ...interface{}) error {
	resp, err := r.Get(uri, values...)
	if err != nil {
		return err
	}
	return json.Unmarshal(resp.body, result)
}

func (r *Req) Post(url string, values ...interface{}) (*Resp, error) {
	r.method = "POST"
	r.url = url
	return r.Do(values...)
}

func (r *Req) PostJson(result interface{}, uri string, values ...interface{}) error {
	resp, err := r.Post(uri, values...)
	if err != nil {
		return err
	}
	return json.Unmarshal(resp.body, result)
}

func (r *Req) Do(values ...interface{}) (resp *Resp, err error) {
	for _, value := range values {
		switch vv := value.(type) {
		case Headers:
			r.headers = make(Headers)
			for k, v := range vv {
				r.headers[k] = v
			}
		case Params:
			r.params = vv
		case url.Values:
			r.values = vv
		case Timeout:
			r.timeout = time.Duration(vv)
		case *tls.Config:
			r.tlsConfig = vv
		default:
			return nil, ErrorParamInvalid
		}
	}

	if err = r.newReq(); err != nil {
		return
	}

	r.newClient()

	return r.do()
}

func (r *Req) newReq() (err error) {
	if r.method == "GET" {
		if r.params == nil && r.values == nil {
			r.req, err = http.NewRequest("GET", r.url, nil)
		} else if r.params != nil {
			r.values = url.Values{}
			for k, v := range r.params {
				r.values.Set(k, fmt.Sprint(v))
			}
		}
		if r.values != nil {
			r.req, err = http.NewRequest("GET", r.url+"?"+r.values.Encode(), nil)
		}
	} else {
		if r.params == nil && r.values == nil {
			r.req, err = http.NewRequest("POST", r.url, nil)
		} else if r.params != nil {
			if r.headers != nil && r.headers["Content-Type"] == "application/x-www-form-urlencoded" {
				r.values = url.Values{}
				for k, v := range r.params {
					r.values.Set(k, fmt.Sprint(v))
				}
			} else {
				// json
				var data []byte
				if data, err = json.Marshal(r.params); err != nil {
					return errors.Wrap(err, "json params failed")
				}
				r.req, err = http.NewRequest("POST", r.url, bytes.NewReader(data))
			}
		}
		if r.values != nil {
			r.req, err = http.NewRequest("POST", r.url, strings.NewReader(r.values.Encode()))
		}
	}
	if err != nil {
		return errors.Wrapf(err, "http.NewRequest %s %s failed", r.method, r.url)
	}

	if r.headers != nil {
		for k, v := range r.headers {
			r.req.Header.Set(k, v)
		}
	}

	return nil
}

func (r *Req) newClient() {
	dialer := &net.Dialer{
		Timeout: r.timeout,
	}
	transport := &http.Transport{
		DialContext:       dialer.DialContext,
		TLSClientConfig:   r.tlsConfig,
		DisableKeepAlives: true,
	}
	r.client = &http.Client{
		Timeout:   r.timeout,
		Transport: transport,
	}
}

func (r *Req) do() (resp *Resp, err error) {
	httpResp, err := r.client.Do(r.req)
	if err != nil {
		return
	}
	defer httpResp.Body.Close()

	data, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return
	}

	resp = &Resp{
		statusCode: httpResp.StatusCode,
		body:       data,
	}
	return
}

func NewReq() *Req {
	return &Req{
		timeout: 10 * time.Second,
	}
}

type Resp struct {
	statusCode int
	body       []byte
}

func (r *Resp) StatusCode() int {
	return r.statusCode
}

func (r *Resp) Bytes() []byte {
	return r.body
}

func (r *Resp) Text() string {
	return string(r.body)
}

func (r *Resp) Json(v interface{}) error {
	return json.Unmarshal(r.body, v)
}

func Get(url string, values ...interface{}) (*Resp, error) {
	return NewReq().Get(url, values...)
}

func GetJson(result interface{}, url string, values ...interface{}) error {
	return NewReq().GetJson(result, url, values...)
}

func Post(url string, values ...interface{}) (*Resp, error) {
	return NewReq().Post(url, values...)
}

func PostJson(result interface{}, url string, values ...interface{}) error {
	return NewReq().PostJson(result, url, values...)
}
