package catalog

import (
	"fmt"
	"sync"
)

var catalog = map[uint32]string{}
var guard sync.RWMutex

func Put(deviceID uint32, oid string) {
	guard.Lock()
	defer guard.Unlock()

	catalog[deviceID] = oid

}

func Get(deviceID uint32) string {
	guard.RLock()
	defer guard.RUnlock()

	oid, ok := catalog[deviceID]
	if !ok {
		item := 0
	loop:
		for {
			item += 1
			oid = fmt.Sprintf("0.1.1.%d", item)
			for _, v := range catalog {
				if v == oid {
					continue loop
				}
			}

			catalog[deviceID] = oid
			break
		}
	}

	return oid
}
