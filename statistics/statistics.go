package statistics

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"pressure-test/base"
	"pressure-test/config"
)

type StdoutResponse struct {
	responseChan chan *base.Response
	concurrent   *uint64
	wg           *sync.WaitGroup
}

func NewStdoutResponse(wg *sync.WaitGroup, ch chan *base.Response, concurrent *uint64) *StdoutResponse {
	return &StdoutResponse{wg: wg, responseChan: ch, concurrent: concurrent}
}

func (s *StdoutResponse) Start() {
	detail := &StdoutDetail{
		StartTime:  time.Now(),
		Concurrent: *s.concurrent,
		CodeMap:    make(map[int]int),
	}
	detail.MinTime = math.MaxUint64
	chanIds := make(map[int]struct{})
	ctx, cancel := context.WithCancel(context.Background())
	s.wg.Add(1)
	go detail.tickerCalculateAndPrint(s.wg, ctx)
	header()
	func() {
		var (
			data *base.Response
			t    = time.NewTimer(time.Second)
		)
		for {
			select {
			case <-t.C:
				if len(s.responseChan) == 0 {
					return
				}
				t.Stop()
			case data = <-s.responseChan:
				if config.Cfg.DebugMode {
					fmt.Println(data)
				}
				mux := RWMuxGet()
				mux.Lock()
				//状态码统计
				detail.CodeMap[data.Code]++
				// 统计并发数
				if _, ok := chanIds[data.ChanId]; !ok {
					chanIds[data.ChanId] = struct{}{}
					detail.ChanIdLen = len(chanIds)
				}
				mux.Unlock()
				mux.Put()
				if data.Code != 200 { //失败统计
					atomic.AddUint64(&detail.FailureNum, 1)
					goto UNLOCK
				}
				//成功统计
				//请求总时间统计
				atomic.AddUint64(&detail.ProcessingTime, uint64(data.TimeCost))
				if detail.MaxTime < uint64(data.TimeCost) {
					atomic.StoreUint64(&detail.MaxTime, uint64(data.TimeCost))
				}
				if detail.MinTime > uint64(data.TimeCost) {
					atomic.StoreUint64(&detail.MinTime, uint64(data.TimeCost))
				}
				atomic.AddUint64(&detail.SuccessNum, 1)
			UNLOCK:
				data.Put()
				t.Reset(time.Second)
			}
		}
	}()
	atomic.StoreUint64(&detail.RequestTime,
		uint64(time.Since(detail.StartTime).Nanoseconds()-time.Second.Nanoseconds()))
	cancel()
	s.wg.Add(1)
	go func(wg *sync.WaitGroup, detail *StdoutDetail) {
		calculateData(detail.Concurrent, detail.ProcessingTime, detail.RequestTime, detail.MaxTime,
			detail.MinTime, detail.SuccessNum, detail.FailureNum, detail.ChanIdLen, detail.CodeMap)
		fmt.Println("\n")
		fmt.Println("*************************  结果 stat  ****************************")
		fmt.Printf("处理协程数量:%v\n", *s.concurrent)
		total := detail.SuccessNum + detail.FailureNum
		fmt.Printf("请求总数:%v 总请求时间:%.4f秒 成功数:%v 失败数:%v 成功率:%.2f\n",
			total, float64(detail.RequestTime)/1e9,
			detail.SuccessNum, detail.FailureNum,
			float64(detail.SuccessNum)*100.0/float64(total))
		fmt.Println("*************************  结果 end   ****************************")
		fmt.Println("\n")
		wg.Done()
	}(s.wg, detail)
}

type StdoutDetail struct {
	StartTime      time.Time //开始时间
	Concurrent     uint64    //当前并发数
	ProcessingTime uint64    // 处理总时间
	RequestTime    uint64    // 请求总时间
	MaxTime        uint64    // 最大时长
	MinTime        uint64    //最小时长
	SuccessNum     uint64
	FailureNum     uint64
	ChanIdLen      int
	CodeMap        map[int]int
}

//定时计算并输出
func (s *StdoutDetail) tickerCalculateAndPrint(wg *sync.WaitGroup, ctx context.Context) {
	// 定时输出一次计算结果
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ticker.C:
			atomic.StoreUint64(&s.RequestTime, uint64(time.Since(s.StartTime).Nanoseconds()))
			go calculateData(s.Concurrent, s.ProcessingTime, s.RequestTime, s.MaxTime,
				s.MinTime, s.SuccessNum, s.FailureNum, s.ChanIdLen, s.CodeMap)
		case <-ctx.Done():
			ticker.Stop()
			wg.Done()
			return
		}
	}
}

// 计算数据
func calculateData(concurrent, processingTime, requestTime, maxTime, minTime, successNum, failureNum uint64, chanIdLen int, codeMap map[int]int) {
	if atomic.LoadUint64(&processingTime) == 0 {
		atomic.StoreUint64(&processingTime, 1)
	}
	var (
		qps              float64
		rps              float64
		averageTime      float64
		maxTimeFloat     float64
		minTimeFloat     float64
		requestTimeFloat float64
	)
	// 平均 每个协程成功数*总协程数据/总耗时 (每秒)   QPS = 总请求数 / ( 进程总数 * 请求时间 )
	qps = float64(successNum*1e9) / float64(requestTime)
	// 平均时长 总耗时/总请求数/并发数 纳秒=>毫秒
	if successNum != 0 {
		averageTime = float64(processingTime) / float64(successNum*1e6)
	}
	rps = float64(concurrent) * 1e3 / averageTime
	// 纳秒=>毫秒
	maxTimeFloat = float64(maxTime) / 1e6
	minTimeFloat = float64(minTime) / 1e6
	requestTimeFloat = float64(requestTime) / 1e9

	// 打印的时长都为毫秒
	// result := fmt.Sprintf("请求总数:%8d|successNum:%8d|failureNum:%8d|qps:%9.3f|maxTime:%9.3f|minTime:%9.3f|平均时长:%9.3f|errCode:%v", successNum+failureNum, successNum, failureNum, qps, maxTimeFloat, minTimeFloat, averageTime, errCode)
	// fmt.Println(result)
	table(successNum, failureNum, codeMap, qps, averageTime, maxTimeFloat, minTimeFloat, requestTimeFloat, rps, chanIdLen)
}

// 打印表头信息
func header() {
	fmt.Printf("\n\n")
	// 打印的时长都为毫秒 总请数
	fmt.Println("─────┬───────┬───────┬───────┬────────┬────────┬────────┬────────┬────────┬────────")
	result := fmt.Sprintf(" 耗时│ 并发数│ 成功数│ 失败数│   QPS  │最长耗时│最短耗时│平均耗时│ 错误码  │ RPS")
	fmt.Println(result)
	// result = fmt.Sprintf("耗时(s)  │总请求数│成功数│失败数│QPS│最长耗时│最短耗时│平均耗时│错误码")
	// fmt.Println(result)
	fmt.Println("─────┼───────┼───────┼───────┼────────┼────────┼────────┼────────┼────────┼────────")
}

// 打印表格
func table(successNum, failureNum uint64, errCode map[int]int, qps, averageTime, maxTimeFloat,
	minTimeFloat, requestTimeFloat, rps float64, chanIdLen int) {
	// 打印的时长都为毫秒
	result := fmt.Sprintf("%4.0fs│%7d│%7d│%7d│%8.2f│%8.2f│%8.2f│%8.2f│%v|%8.2f",
		requestTimeFloat, chanIdLen, successNum, failureNum, qps, maxTimeFloat, minTimeFloat, averageTime,
		printMap(errCode), rps)
	fmt.Println(result)
}

// 输出错误码、次数 节约字符(终端一行字符大小有限)
var bufPool = sync.Pool{New: func() interface{} {
	return new(bytes.Buffer)
}}

func printMap(errCode map[int]int) string {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	for key, value := range errCode {
		buf.WriteString(fmt.Sprintf("%d:%d", key, value))
		buf.WriteString(";")
	}
	return buf.String()
}
