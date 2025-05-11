package ratelimiter

import (
	"net"
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   int       // максимальное количество токенов
	tokens     int       // текущее количество токенов
	rate       int       // токенов в секунду
	lastUpdate time.Time // последнее обновление
	mutex      sync.Mutex
}

type RateLimiter struct {
	buckets  map[string]*TokenBucket // ключ — IP клиента
	mutex    sync.Mutex
	capacity int
	rate     int
}

func NewRateLimiter(capacity, rate int) *RateLimiter {
	return &RateLimiter{
		buckets:  make(map[string]*TokenBucket),
		capacity: capacity,
		rate:     rate,
	}
}

func (rl *RateLimiter) getBucket(key string) *TokenBucket {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = &TokenBucket{
			capacity:   rl.capacity,
			tokens:     rl.capacity,
			rate:       rl.rate,
			lastUpdate: time.Now(),
		}
		rl.buckets[key] = bucket
	}
	return bucket
}

func (rl *RateLimiter) Allow(ip net.IP) bool {
	bucket := rl.getBucket(ip.String())
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastUpdate).Seconds()
	// Пополняем токены
	newTokens := int(elapsed * float64(bucket.rate))
	if newTokens > 0 {
		bucket.tokens = min(bucket.capacity, bucket.tokens+newTokens)
		bucket.lastUpdate = now
	}
	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}
	return false
}
