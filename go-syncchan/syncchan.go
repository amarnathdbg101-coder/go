package syncchan

import (
	"context"
	"sync"
)

// OrDone wraps a readable channel with context cancellation, preventing reader goroutine leaks.
func OrDone[T any](ctx context.Context, in <-chan T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case val, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- val:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

// FanIn merges multiple input channels of type T into a single consolidated output channel.
// Closes the output channel when all input channels are closed or when ctx is cancelled.
func FanIn[T any](ctx context.Context, chans ...<-chan T) <-chan T {
	out := make(chan T)
	if len(chans) == 0 {
		close(out)
		return out
	}

	var wg sync.WaitGroup
	wg.Add(len(chans))

	for _, ch := range chans {
		go func(c <-chan T) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case val, ok := <-c:
					if !ok {
						return
					}
					select {
					case out <- val:
					case <-ctx.Done():
						return
					}
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// Tee duplicates an input channel into two independent identical output channels.
func Tee[T any](ctx context.Context, in <-chan T) (<-chan T, <-chan T) {
	out1 := make(chan T)
	out2 := make(chan T)

	go func() {
		defer close(out1)
		defer close(out2)

		for val := range OrDone(ctx, in) {
			// Write to both channels without blocking one on the other
			var o1, o2 = out1, out2
			for i := 0; i < 2; i++ {
				select {
				case <-ctx.Done():
					return
				case o1 <- val:
					o1 = nil
				case o2 <- val:
					o2 = nil
				}
			}
		}
	}()

	return out1, out2
}

// FanOut duplicates values from in to n separate subscriber channels.
func FanOut[T any](ctx context.Context, in <-chan T, n int) []<-chan T {
	if n <= 0 {
		return nil
	}

	outs := make([]chan T, n)
	readOuts := make([]<-chan T, n)
	for i := 0; i < n; i++ {
		outs[i] = make(chan T)
		readOuts[i] = outs[i]
	}

	go func() {
		defer func() {
			for _, ch := range outs {
				close(ch)
			}
		}()

		for val := range OrDone(ctx, in) {
			for _, ch := range outs {
				select {
				case <-ctx.Done():
					return
				case ch <- val:
				}
			}
		}
	}()

	return readOuts
}

// Take reads at most n values from in and closes the returned channel.
func Take[T any](ctx context.Context, in <-chan T, n int) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		if n <= 0 {
			return
		}
		count := 0
		for val := range OrDone(ctx, in) {
			select {
			case <-ctx.Done():
				return
			case out <- val:
				count++
				if count >= n {
					return
				}
			}
		}
	}()
	return out
}

// Drop skips the first n values from in and forwards all subsequent values.
func Drop[T any](ctx context.Context, in <-chan T, n int) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		dropped := 0
		for val := range OrDone(ctx, in) {
			if dropped < n {
				dropped++
				continue
			}
			select {
			case <-ctx.Done():
				return
			case out <- val:
			}
		}
	}()
	return out
}

// Buffer introduces an intermediate buffered channel between in and out.
func Buffer[T any](ctx context.Context, in <-chan T, size int) <-chan T {
	if size <= 0 {
		size = 1
	}
	out := make(chan T, size)
	go func() {
		defer close(out)
		for val := range OrDone(ctx, in) {
			select {
			case <-ctx.Done():
				return
			case out <- val:
			}
		}
	}()
	return out
}

// Collect drains an input channel into a slice until it is closed.
func Collect[T any](in <-chan T) []T {
	var result []T
	for val := range in {
		result = append(result, val)
	}
	return result
}
