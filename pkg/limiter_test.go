package pkg_test

import (
	"sync"
	"testing"
	"time"

	"github.com/danielbintar/ratelimit/pkg"
)

func TestHit(t *testing.T) {
	var wg sync.WaitGroup
	const workerCount = 1000
	const hitEveryWorker = 100
	wg.Add(workerCount)
	now, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:10Z")
	duration := time.Minute
	limiter := pkg.NewLimiter()
	defer limiter.Close()
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < hitEveryWorker; j++ {
				limiter.Hit("a", duration, now)
			}
		}()
	}

	wg.Wait()

	keyHit := limiter.Hit("a", duration, now.Add(3*time.Second))
	expectedKeyHit := workerCount*hitEveryWorker + 1
	if keyHit != expectedKeyHit {
		t.Fatalf("Hit key should be %d, not %d", expectedKeyHit, keyHit)
	}

	laterKeyHit := limiter.Hit("a", duration, now.Add(51*time.Second))
	if laterKeyHit != 1 {
		t.Fatalf("Hit later key should be 1, not %d", laterKeyHit)
	}

	sameMinuteHit := limiter.Hit("a", duration, now.Add(1*time.Hour))
	if sameMinuteHit != 1 {
		t.Fatalf("Hit same minute key should be 1, not %d", sameMinuteHit)
	}

	newKeyHit := limiter.Hit("b", duration, now)
	if newKeyHit != 1 {
		t.Fatalf("Hit new key should be 1, not %d", newKeyHit)
	}
}
