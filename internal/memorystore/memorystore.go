package memorystore

import (
	"errors"
	"sync"
)

var mu sync.Mutex // guards memoryStore
var store [1000000]bool
var storeLen = 0
var initialized = false

func Init() {
	if initialized {
		panic("MemoryStore Init was called more than once")
	}

	//todo pull store size from config files
	storeLen = len(store)
	initialized = true
}

func DoCheck(checkboxNbr int, checked bool) error {
	if checkboxNbr < 0 || checkboxNbr >= storeLen {
		return errors.New("Invalid CheckboxNumber")
	}
	mu.Lock()
	store[checkboxNbr] = checked
	mu.Unlock()

	return nil
}
