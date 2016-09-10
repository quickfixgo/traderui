package internal

import (
	"strconv"
	"sync"
)

type BasicClOrdIDFactory struct {
	clOrdIDLock sync.Mutex
	clOrdID     int
}

func (f *BasicClOrdIDFactory) NextClOrdID() string {
	f.clOrdIDLock.Lock()
	defer f.clOrdIDLock.Unlock()

	f.clOrdID++
	return strconv.Itoa(f.clOrdID)
}
