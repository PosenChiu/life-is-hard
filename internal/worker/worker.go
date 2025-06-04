package worker

import "sync"

// Task represents a unit of work executed by the pool.
type Task func()

// Pool defines a simple worker pool.
type Pool interface {
	Submit(Task)
	Stop()
}

// NewPool creates a pool with n workers. n<=0 defaults to 1.
func NewPool(n int) Pool {
	if n <= 0 {
		n = 1
	}
	p := &pool{jobs: make(chan Task)}
	p.wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer p.wg.Done()
			for job := range p.jobs {
				if job != nil {
					job()
				}
			}
		}()
	}
	return p
}

type pool struct {
	jobs chan Task
	wg   sync.WaitGroup
}

func (p *pool) Submit(t Task) {
	p.jobs <- t
}

func (p *pool) Stop() {
	close(p.jobs)
	p.wg.Wait()
}
