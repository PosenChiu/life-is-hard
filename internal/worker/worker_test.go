package worker

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPool(t *testing.T) {
	p := NewPool(3)
	var mu sync.Mutex
	count := 0
	for i := 0; i < 5; i++ {
		p.Submit(func() {
			mu.Lock()
			count++
			mu.Unlock()
		})
	}
	p.Stop()
	require.Equal(t, 5, count)
}
