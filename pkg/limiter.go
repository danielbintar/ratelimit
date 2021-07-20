package pkg

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type node struct {
	key string
	hit int

	expiration time.Time
}

type Limiter struct {
	sync.RWMutex

	data     []node
	keyIndex map[string]int

	closeSign chan struct{}
}

func NewLimiter() *Limiter {
	l := &Limiter{
		data:      make([]node, 0),
		keyIndex:  make(map[string]int),
		closeSign: make(chan struct{}),
	}

	go l.periodicallyCleanup()

	return l
}

func (this *Limiter) Close() {
	this.closeSign <- struct{}{}
}

func (this *Limiter) cleanup() {
	this.Lock()
	defer this.Unlock()

	n := len(this.data)
	for n > 0 {
		id := rand.Intn(n)
		if this.data[id].expiration.After(time.Now()) {
			break
		}

		delete(this.keyIndex, this.data[id].key)
		this.keyIndex[this.data[n-1].key] = id
		this.data[id] = this.data[n-1]
		this.data = this.data[:n-1]

		n--
	}
}

func (this *Limiter) periodicallyCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-this.closeSign:
			return
		case <-ticker.C:
			this.cleanup()
		}
	}
}

func (this *Limiter) Hit(key string, duration time.Duration, now time.Time) int {
	this.Lock()
	defer this.Unlock()

	floor := now.Round(duration)
	key = fmt.Sprintf("%s:%d", key, floor.Unix())

	i, found := this.keyIndex[key]
	if !found {
		this.data = append(this.data, node{
			key:        key,
			hit:        0,
			expiration: floor.Add(duration),
		})
		i = len(this.data) - 1
		this.keyIndex[key] = i
	}

	this.data[i].hit++
	return this.data[i].hit
}
