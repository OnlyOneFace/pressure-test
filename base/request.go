package base

import (
	"context"
	"sync"
)

//模型接口
type Transaction interface {
	InteractionChan(chan *Response)
	SetContext(ctx context.Context)
	Preset()
	Do()
	Concurrent() *uint64
}

//事务模型
type (
	TransactionInstance struct {
		Data
		ConcurrentNum uint64
		responseChan  chan *Response
		ctx           context.Context
		wg            *sync.WaitGroup
	}

	Data struct {
		Id                 string
		Name               string
		PresetTransactions []*PresetCase //预置方案
		DoTransactions     []*Case       //常规方案
	}
)

func (t *TransactionInstance) InteractionChan(ch chan *Response) {
	t.responseChan = ch
}

func (t *TransactionInstance) SetContext(ctx context.Context) {
	t.ctx = ctx
}

func (t *TransactionInstance) Preset() {
	for _, presetTransaction := range t.PresetTransactions {
		for _, client := range presetTransaction.Clients {
			t.responseChan <- client.Exec()
		}
	}
}

func (t *TransactionInstance) Do() {
	t.wg = new(sync.WaitGroup)
	for _, doTransaction := range t.DoTransactions {
		doTransaction.RunMode.Run(t.ctx, t.wg, &t.ConcurrentNum, func(concurrentNo int, count int) {
			for _, client := range doTransaction.Clients {
				resp := client.Exec()
				resp.SetId(concurrentNo, count)
				t.responseChan <- resp
			}
		})
	}
	t.wg.Wait()
}

func (t *TransactionInstance) Concurrent() *uint64 {
	return &t.ConcurrentNum
}

type (
	//常规用例
	Case struct {
		Clients []Client
		RunMode RunMode
	}

	//预置用例
	PresetCase struct {
		Clients []Client
	}

	Client interface {
		Exec() *Response
	}
)
