package pkg

import (
	"fmt"
	"sync"
	"time"
)

type node struct {
	key string
	hit int
}

type Limiter struct {
	sync.RWMutex

	data     []node
	keyIndex map[string]int
}

func NewLimiter() Limiter {
	return Limiter{
		data:     make([]node, 0),
		keyIndex: make(map[string]int),
	}
}

func (this *Limiter) Hit(key string, duration time.Duration, now time.Time) int {
	this.Lock()
	defer this.Unlock()

	key = fmt.Sprintf("%s:%d", key, now.Round(duration).Unix())

	i, found := this.keyIndex[key]
	if !found {
		this.data = append(this.data, node{
			key: key,
			hit: 0,
		})
		i = len(this.data) - 1
		this.keyIndex[key] = i
	}

	this.data[i].hit++
	return this.data[i].hit
}
