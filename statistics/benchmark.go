package statistics

import (
	"context"
	"sync"

	"pressure-test/base"
)

type CommitResponse interface {
	Start()
}

func ReceivingResponse(crs ...CommitResponse) {
	for _, cr := range crs {
		go cr.Start()
	}
}

type (
	BenchMark struct {
		TotalRequest int64
		TotalTime    int64
		MaxTime      int64
		AvgTime      int64
		TotalReceive int64
		TotalSend    int64
		AWSuccess    int64
		AWFailed     int64
		TP50         int64
		TP90         int64
		TP99         int64
		TakenTime    int64
		RPS          float64
	}

	BenchMarkInstance struct {
		BenchMarkPool *sync.Pool
		ctx           context.Context
		responseChan  chan *base.Response
	}
)

func NewBenchMarkInstance(ctx context.Context, ch chan *base.Response) *BenchMarkInstance {
	return &BenchMarkInstance{
		BenchMarkPool: &sync.Pool{
			New: func() interface{} {
				return new(BenchMark)
			}},
		ctx:          ctx,
		responseChan: ch,
	}
}

func (b *BenchMarkInstance) Start() {
	var (
		resp *base.Response
		bm   *BenchMark
	)
	for {
		select {
		case <-b.ctx.Done():
			return
		case resp = <-b.responseChan:
			bm = b.get()
			bm.TotalRequest++
			bm.TotalReceive += resp.ContentLength
			if resp.Code != 0 {
				bm.AWSuccess++
				bm.TotalTime += resp.TimeCost
				if bm.MaxTime < resp.TimeCost {
					bm.MaxTime = resp.TimeCost
				}
				break
			}
			bm.AWFailed++
			resp.Put()
			b.put(bm)
		}
	}
}

func (b *BenchMarkInstance) get() *BenchMark {
	for {
		r := b.BenchMarkPool.Get()
		resp, ok := r.(*BenchMark)
		if !ok {
			continue
		}
		return resp
	}
}

func (b *BenchMarkInstance) put(bm *BenchMark) {
	bm.TotalRequest = 0
	bm.TotalTime = 0
	bm.MaxTime = 0
	bm.AvgTime = 0
	bm.TotalReceive = 0
	bm.TotalSend = 0
	bm.AWSuccess = 0
	bm.AWFailed = 0
	bm.TP50 = 0
	bm.TP90 = 0
	bm.TP99 = 0
	bm.TakenTime = 0
	bm.RPS = 0
	b.BenchMarkPool.Put(bm)
}
