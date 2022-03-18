//+build !windows

package main

import (
	"sync"
	"syscall"
)

var umask int
var mu sync.Mutex

func SetUmask(mask int) {
	mu.Lock()
	defer mu.Unlock()
	umask = syscall.Umask(0)
}

func UnsetUmask() {
	mu.Lock()
	defer mu.Unlock()
	syscall.Umask(umask)
}
