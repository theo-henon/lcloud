package protocols

import (
	"net"
	"sync"
	"time"
)

const authFailureLimit = 5

type authRateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

func newAuthRateLimiter() *authRateLimiter {
	return &authRateLimiter{attempts: make(map[string][]time.Time)}
}

func (l *authRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-time.Minute)
	prev := l.attempts[ip]
	filtered := prev[:0]
	for _, ts := range prev {
		if ts.After(cutoff) {
			filtered = append(filtered, ts)
		}
	}
	if len(filtered) >= authFailureLimit {
		l.attempts[ip] = filtered
		return false
	}
	l.attempts[ip] = filtered
	return true
}

func (l *authRateLimiter) recordFailure(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.attempts[ip] = append(l.attempts[ip], time.Now())
}

func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}
