package main

import "sync"

type Counter struct {
	mu    sync.Mutex
	count int
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) Inc() {
	c.count++
}

func (c *Counter) Value() int {
	return c.count
}
