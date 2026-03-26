package workspace

import "sync"

type Locker struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func NewLocker() *Locker {
	return &Locker{
		locks: map[string]*sync.Mutex{},
	}
}

func (l *Locker) Acquire(key string) func() {
	l.mu.Lock()
	lock, ok := l.locks[key]
	if !ok {
		lock = &sync.Mutex{}
		l.locks[key] = lock
	}
	l.mu.Unlock()

	lock.Lock()
	return func() {
		lock.Unlock()
	}
}
