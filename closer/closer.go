package closer

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Closer - структура, которая выполняет graceful shutdown
type Closer struct {
	mu    sync.Mutex
	funcs []Func
}

// Add добавляет функцию завершения в Closer
func (c *Closer) Add(f Func) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.funcs = append(c.funcs, f)
}

// Close совершает Graceful shutdown
func (c *Closer) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	messages := make([]string, 0, len(c.funcs))
	complete := make(chan struct{}, 1)

	go func() {
		for _, f := range c.funcs {
			if err := f(ctx); err != nil {
				messages = append(messages, fmt.Sprintf("[!] %v", err))
			}
		}
		complete <- struct{}{}
	}()

	select {
	case <-complete:
		break
	case <-ctx.Done():
		return fmt.Errorf("shutdown cancelled: %v", ctx.Err())
	}

	if len(messages) > 0 {
		return fmt.Errorf(
			"shutdown finished with errors: \n%s", strings.Join(messages, "\n"),
		)
	}
	return nil
}

// Func - тип функций, который принимает Closer
type Func func(ctx context.Context) error
