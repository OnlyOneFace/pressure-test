package base

import (
	"strconv"
	"sync"
)

type Response struct {
	Id            string
	ChanId        int
	Header        map[string][]string
	Body          []byte
	Code          int
	ContentLength int64
	TimeCost      int64 //us
	ErrInfo       error
}

func (r *Response) SetId(concurrentNo, roundNo int) {
	r.Id = strconv.Itoa(concurrentNo) + "_" + strconv.Itoa(roundNo)
	r.ChanId = concurrentNo
}

func (r *Response) Reset() {
	r.Id = ""
	r.ChanId = 0
	r.Header = make(map[string][]string)
	r.Body = []byte{}
	r.Code = 0
	r.ContentLength = 0
	r.TimeCost = 0
	r.ErrInfo = nil
}

func (r *Response) Put() {
	r.Reset()
	responsePool.Put(r)
}

var responsePool = sync.Pool{New: func() interface{} {
	return &Response{
		Id:            "",
		ChanId:        0,
		Header:        make(map[string][]string),
		Body:          []byte{},
		Code:          0,
		ContentLength: 0,
		TimeCost:      0,
		ErrInfo:       nil,
	}
}}

func ResponseGet() *Response {
	for {
		r := responsePool.Get()
		resp, ok := r.(*Response)
		if !ok {
			continue
		}
		return resp
	}
}
