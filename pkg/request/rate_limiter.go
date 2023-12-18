package request

import (
	"golang.org/x/time/rate"
)

type RateLimiter interface {
	// Allow returns true if the request is allowed.
	Allow(key string) bool
}

type rateLimiterImpl struct {
	// limiter is the limiter.
	limiters map[string]*rate.Limiter
}

func NewRateLimiter() RateLimiter {
	return &rateLimiterImpl{
		limiters: make(map[string]*rate.Limiter),
	}
}

func (r *rateLimiterImpl) Allow(key string) bool {
	// Rate limits the request.
	limiter, ok := r.limiters[key]
	if !ok {
		limiter = rate.NewLimiter(10, 5)
		r.limiters[key] = limiter
	}

	return limiter.Allow()
}
