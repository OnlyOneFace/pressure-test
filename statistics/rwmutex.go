package statistics

import "sync"

type myRWMutex struct {
	*sync.RWMutex
}

var myRWMutexPool = sync.Pool{New: func() interface{} {
	return &myRWMutex{new(sync.RWMutex)}
}}

func (m *myRWMutex) Put() {
	myRWMutexPool.Put(m)
}

func RWMuxGet() *myRWMutex {
	return myRWMutexPool.Get().(*myRWMutex)
}
