package base

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

//标准http
type StandardHttp struct {
	Url     string            //url
	Method  string            //方法
	Header  map[string]string //请求头
	TimeOut time.Duration     //请求超时时间
	Body    io.Reader
}

var (
	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 跳过证书验证
		}}

	fastClient = fasthttp.Client{
		TLSConfig:           &tls.Config{InsecureSkipVerify: true},
		ReadTimeout:         500 * time.Millisecond,
		MaxConnsPerHost:     60000,
		MaxIdleConnDuration: 10 * time.Second,
	}
)

func (s *StandardHttp) Exec() *Response {
	response := ResponseGet()
	httpClient.Timeout = s.TimeOut
	req, err := http.NewRequest(strings.ToUpper(s.Method), s.Url, s.Body)
	if err != nil {
		response.ErrInfo = err
		return response
	}
	for key, value := range s.Header {
		req.Header.Set(key, value)
	}
	var (
		resp  *http.Response
		start = time.Now()
	)
	if resp, err = httpClient.Do(req); err != nil {
		response.ErrInfo = err
		return response
	}
	defer resp.Body.Close()
	response.TimeCost = time.Since(start).Nanoseconds()
	response.Code = resp.StatusCode
	response.Header = resp.Header
	if response.Body, err = ioutil.ReadAll(resp.Body); err != nil {
		response.ErrInfo = err
		return response
	}
	response.ContentLength = int64(len(response.Body))
	return response
}

type FastHttp struct {
	Url     string            //url
	Method  string            //方法
	Header  map[string]string //请求头
	TimeOut time.Duration     //请求超时时间
	Body    fasthttp.StreamWriter
}

func (f *FastHttp) Exec() *Response {
	response := ResponseGet()
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()
	for key, value := range f.Header {
		req.Header.Set(key, value)
	}
	req.Header.SetMethod(f.Method)
	req.SetRequestURI(f.Url)
	if f.Body != nil {
		req.SetBodyStreamWriter(f.Body)
	}
	fastClient.ReadTimeout = f.TimeOut
	var (
		err   error
		start = time.Now()
	)
	if f.TimeOut == 0 {
		err = fastClient.Do(req, resp)
	} else {
		err = fastClient.DoTimeout(req, resp, f.TimeOut)
	}
	if err != nil {
		response.ErrInfo = err
		return response
	}
	response.TimeCost = time.Since(start).Nanoseconds()
	resp.Header.VisitAll(func(key, value []byte) {
		response.Header[string(key)] = append(response.Header[string(key)], string(value))
	})
	response.Code = resp.Header.StatusCode()
	response.Body = resp.Body()
	response.ContentLength = int64(resp.Header.ContentLength())
	return response
}
