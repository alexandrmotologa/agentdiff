package main

import (
	"sync"
	"testing"
)

func TestConcurrentIncrements(t *testing.T) {
	c := NewCounter()
	var wg sync.WaitGroup
	workers := 50
	increments := 100

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < increments; j++ {
				c.Inc()
			}
		}()
	}

	wg.Wait()
	expected := workers * increments
	if c.Value() != expected {
		t.Fatalf("expected count %d, got %d", expected, c.Value())
	}
}
