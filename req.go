package req

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"time"
)

type Header map[string]string
type Map map[string]interface{}
type Param map[string]interface{}
type Data map[string]interface{}
type StatusCode int
type Debug bool
type Timeout time.Duration
type Retry int

type List []interface{}

const (
	OK      = StatusCode(200)
	Created = StatusCode(201)
	DebugOn = Debug(true)
	Retry3  = Retry(3)
	Retry10 = Retry(10)
)

var ErrorParamInvalid = errors.New("req param invalid")

type Req struct {
	statusCode StatusCode
	sa         *gorequest.SuperAgent
}

func (r *Req) Get(url string, values ...interface{}) (*Resp, error) {
	r.sa = gorequest.New().Get(url)
	return r.Do(values...)
}

func (r *Req) GetJson(result interface{}, url string, values ...interface{}) error {
	resp, err := r.Get(url, values...)
	if err != nil {
		return err
	}
	return json.Unmarshal(resp.body, result)
}

func (r *Req) Post(url string, values ...interface{}) (*Resp, error) {
	r.sa = gorequest.New().Post(url)
	return r.Do(values...)
}

func (r *Req) PostJson(result interface{}, url string, values ...interface{}) error {
	resp, err := r.Post(url, values...)
	if err != nil {
		return err
	}
	return json.Unmarshal(resp.body, result)
}

func (r *Req) Do(values ...interface{}) (resp *Resp, err error) {
	retryCount := 1
	var timeout bool
	for _, value := range values {
		switch vv := value.(type) {
		case Header:
			for k, v := range vv {
				r.sa.Set(k, v)
			}
		case Param:
			for k, v := range vv {
				r.sa.Param(k, fmt.Sprint(v))
			}
		//case Data:
		//	r.sa.Send(vv)
		case StatusCode:
			r.statusCode = vv
		case Debug:
			r.sa.SetDebug(bool(vv))
		case Timeout:
			timeout = true
			r.sa.Timeout(time.Duration(vv))
		case Retry:
			retryCount = int(vv)
		case *tls.Config:
			r.sa.TLSClientConfig(vv)
		default:
			//return nil, ErrorParamInvalid
			r.sa.Send(vv)
		}
	}

	//若没有设置超时，则设置默认超时
	if !timeout {
		r.sa.Timeout(time.Second * 30)
	}

	for i := 0; i < retryCount; i++ {
		httpResp, body, errs := r.sa.EndBytes()
		if len(errs) > 0 {
			err = fmt.Errorf("req Do error: %v", errs[0])
			continue
		}

		if httpResp.StatusCode != int(r.statusCode) {
			err = fmt.Errorf("response status: %s", httpResp.Status)
			continue
		}

		return &Resp{
			body: body,
		}, nil
	}

	return
}

func NewReq() *Req {
	return &Req{statusCode: OK}
}

type Resp struct {
	body []byte
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
