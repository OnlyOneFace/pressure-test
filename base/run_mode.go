package base

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type (
	//运行模式接口
	RunMode interface {
		Run(context.Context, *sync.WaitGroup, *uint64, func(int, int))
	}

	DurationMode struct {
		GradientModes []*GradientMode
	}

	GradientMode struct {
		Num          int           //新增并发数
		DurationTime time.Duration //持续时间单位分钟
	}
)

func (d *DurationMode) Run(ctx context.Context, wg *sync.WaitGroup, nowConCurrent *uint64, f func(int, int)) {
	var (
		timer = time.NewTimer(0)
		i     int
	)
	for _, gradientMode := range d.GradientModes {
		timer.Reset(gradientMode.DurationTime)
		for i = 0; i < gradientMode.Num; i++ {
			atomic.AddUint64(nowConCurrent, 1)
			go func(tempCtx context.Context, concurrentNo int) {
				var count int
				for {
					select {
					case <-tempCtx.Done():
						return
					default:
						count++
						f(concurrentNo, count)
					}
				}
			}(ctx, int(*nowConCurrent))
		}
		<-timer.C
	}
	timer.Stop()
}

type RoundMode struct {
	Num   int //并发数
	Count int //执行次数
}

func (r *RoundMode) Run(ctx context.Context, wg *sync.WaitGroup, u *uint64, f func(int, int)) {
	for i := 0; i < r.Num; i++ {
		atomic.AddUint64(u, 1)
		wg.Add(1)
		go func(tempWg *sync.WaitGroup, concurrentNo int) {
			for j := 0; j < r.Count; j++ {
				f(concurrentNo, j)
			}
			tempWg.Done()
		}(wg, int(*u))
	}
}
