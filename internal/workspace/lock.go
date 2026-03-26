package workspace

import (
	"sync"
	"time"
)

type Locker struct {
	mu        sync.Mutex
	locks     map[string]*lockEntry
	lastClean time.Time
}

type lockEntry struct {
	mu       *sync.Mutex
	lastUsed time.Time
}

func NewLocker() *Locker {
	return &Locker{
		locks:     map[string]*lockEntry{},
		lastClean: time.Now(),
	}
}

func (l *Locker) Acquire(key string) func() {
	l.mu.Lock()
	entry, ok := l.locks[key]
	if !ok {
		entry = &lockEntry{
			mu:       &sync.Mutex{},
			lastUsed: time.Now(),
		}
		l.locks[key] = entry
	}
	entry.lastUsed = time.Now()
	l.cleanIfNeeded()
	l.mu.Unlock()

	entry.mu.Lock()
	return func() {
		entry.mu.Unlock()
	}
}

func (l *Locker) cleanIfNeeded() {
	if time.Since(l.lastClean) < 10*time.Minute {
		return
	}
	l.lastClean = time.Now()
	threshold := time.Now().Add(-30 * time.Minute)
	for key, entry := range l.locks {
		if entry.lastUsed.Before(threshold) {
			delete(l.locks, key)
		}
	}
}
