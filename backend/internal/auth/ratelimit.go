package auth

import (
	"sync"
	"time"
)

const loginLimit = 10

type LoginLimiter struct {
	mu   sync.Mutex
	hits map[string][]time.Time
}

func NewLoginLimiter() *LoginLimiter {
	return &LoginLimiter{hits: map[string][]time.Time{}}
}

func (l *LoginLimiter) Allow(ip, tenantID, email string) bool {
	key := ip + "\x00" + tenantID + "\x00" + email
	now := time.Now()
	window := now.Add(-time.Minute)
	l.mu.Lock()
	defer l.mu.Unlock()
	kept := l.hits[key][:0]
	for _, t := range l.hits[key] {
		if t.After(window) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= loginLimit {
		l.hits[key] = kept
		return false
	}
	l.hits[key] = append(kept, now)
	return true
}
