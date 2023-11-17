package proc

import (
	"os"
	"sync"
)

var (
	envs    = make(map[string]string)
	envLock sync.RWMutex
)

// Env returns the value of the given environment variable.
func Env(name string) string {
	envLock.RLock()
	val, ok := envs[name]
	envLock.RUnlock()

	if ok {
		return val
	}

	val = os.Getenv(name)
	envLock.Lock()
	envs[name] = val
	envLock.Unlock()

	return val
}
