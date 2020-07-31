package base

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

type StandardTcp struct {
	Host           string
	Port           string
	ConnectTimeOut time.Duration
	ReadTimeOut    time.Duration
	body           []byte
	End            interface{} //长度和结束符
}

var tcpDialer = &fasthttp.TCPDialer{Concurrency: 60000}

func (s *StandardTcp) Exec() *Response {
	var (
		writeOnce = true
		response  = ResponseGet()
		err       error
		conn      net.Conn
	)
	if s.ConnectTimeOut != 0 {
		conn, err = tcpDialer.Dial(s.Host + ":" + s.Port)
	} else {
		conn, err = tcpDialer.DialTimeout(s.Host+":"+s.Port, s.ConnectTimeOut)
	}
	if err != nil {
		response.ErrInfo = err
		return response
	}
	if _, err = conn.Write(s.body); err != nil {
		response.ErrInfo = err
		return response
	}
	var (
		fragment   = make([]byte, 1024)
		readLength int
		start      = time.Now()
	)
	for {
		if err = conn.SetReadDeadline(time.Now().Add(s.ReadTimeOut)); err != nil {
			response.ErrInfo = err
			return response
		}
		if readLength, err = conn.Read(fragment); err != nil {
			response.ErrInfo = err
			return response
		}
		if writeOnce {
			response.TimeCost = time.Since(start).Nanoseconds()
			writeOnce = false
		}
		response.Body = append(response.Body, fragment[:readLength]...)
		switch end := s.End.(type) {
		case int:
			if len(response.Body) >= end {
				response.Body = response.Body[:end]
				return response
			}
		case string:
			if index := strings.Index(string(fragment[:readLength]), end); index != -1 {
				response.Body = response.Body[:len(response.Body)-readLength+index]
				return response
			}
		default:
			response.ErrInfo = fmt.Errorf("end(%v) type  unsupport", end)
			return response
		}
	}
}
