package basic

import (
	"strconv"
	"sync"
)

type ClOrdIDGenerator struct {
	clOrdIDLock sync.Mutex
	clOrdID     int
}

func (f *ClOrdIDGenerator) Next() string {
	f.clOrdIDLock.Lock()
	defer f.clOrdIDLock.Unlock()

	f.clOrdID++
	return strconv.Itoa(f.clOrdID)
}
