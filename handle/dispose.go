package handle

import (
	"context"
	"fmt"
	"sync"

	"pressure-test/base"
	"pressure-test/config"
	"pressure-test/handle/parse"
	"pressure-test/statistics"
)

func StartBenchmark() {
	t := parse.GetTransaction()
	if config.Cfg.DebugMode {
		for _, doTransaction := range t.DoTransactions {
			doTransaction.RunMode = &base.RoundMode{Num: 1, Count: 1}
		}
	}
	ExecTransaction(t) //模板生成)
}

func ExecTransaction(transaction base.Transaction) {
	fmt.Println("\n 开始启动!")
	//获取交互的管道
	ch := make(chan *base.Response, int32(655360))
	wg := new(sync.WaitGroup)
	transaction.InteractionChan(ch)
	//起一个goroutine 计算数据
	ctx, cancel := context.WithCancel(context.Background())
	transaction.SetContext(ctx)
	//开启数据收集
	statistics.ReceivingResponse(statistics.NewStdoutResponse(wg, ch, transaction.Concurrent()))
	//执行预置方案
	transaction.Preset()
	//执行常规方案
	transaction.Do()
	// 数据全部处理完成了
	cancel()
	wg.Wait()
	close(ch)
}
