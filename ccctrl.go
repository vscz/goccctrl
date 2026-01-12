package goccctrl

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"
)

// Weighted target must implement this interface
type Weighted interface {
	Weight() int
}

// Validatable result must implement this interface
type Validatable interface {
	IsValid() bool
}

// NoWeighted is a struct that implements the Weighted interface
type NoWeighted struct {
}

func (n NoWeighted) Weight() int {
	return 1
}

type Result[R Validatable] struct {
	Val       R
	Err       error
	MoreValCh <-chan Result[R]
}

// ReqFunc is the generic request function type
type ReqFunc[T Weighted, R Validatable] func(ctx context.Context, target T) Result[R]

type ReqTime struct {
	TotalTimeout time.Duration
	FirstWait    time.Duration
	MinWait      time.Duration
}

func (r *ReqTime) NextWaitTime(wait time.Duration) time.Duration {
	if wait > r.MinWait {
		return wait / 2
	}
	return r.MinWait
}

func NewReqParam[T Weighted, R Validatable](targets []T, reqTime ReqTime, reqFunc ReqFunc[T, R]) *ReqParam[T, R] {
	return &ReqParam[T, R]{
		ReqTime: reqTime,
		Targets: targets,
		ReqFunc: reqFunc,
	}
}

type ReqParam[T Weighted, R Validatable] struct {
	ReqTime ReqTime

	Targets        []T
	ReqFunc        ReqFunc[T, R]
	ShuffleTargets bool
}

// SetShuffleTargets sets whether to shuffle targets
func (r *ReqParam[T, R]) SetShuffleTargets(shuffle bool) {
	r.ShuffleTargets = shuffle
}

// DoShuffleTargets shuffles targets
func (r *ReqParam[T, R]) DoShuffleTargets() {
	if !r.ShuffleTargets {
		return
	}

	newTargets := make([]T, 0, len(r.Targets))
	used := make(map[int]bool)

	for len(used) < len(r.Targets) {
		// weighted pick
		idx := weightedPick(r.Targets, used)
		if idx == -1 {
			break
		}
		used[idx] = true

		newTargets = append(newTargets, r.Targets[idx])

		// if last target, no need to wait
		if len(used) == len(r.Targets) {
			break
		}
	}

	r.Targets = newTargets
}

// Do executes requests progressively with weighted scheduling
func (r *ReqParam[T, R]) Do(ctx context.Context) (res Result[R]) {
	var zero R
	if len(r.Targets) == 0 {
		return Result[R]{Val: zero, Err: errors.New("no targets provided")}
	}

	ctx, _ = context.WithTimeout(ctx, r.ReqTime.TotalTimeout)
	// defer cancel()

	resultCh := make(chan Result[R], len(r.Targets)-1)

	wg := sync.WaitGroup{}
	// check if close resultCh when all requests are done
	defer func() {
		res.MoreValCh = resultCh

		go func() {
			wg.Wait()
			close(resultCh)
		}()
	}()

	wait := r.ReqTime.FirstWait
	r.DoShuffleTargets()
	for idx := range r.Targets {
		wg.Add(1)
		go r.doRequest(ctx, r.Targets[idx], resultCh, &wg)

		// check next by last result or wait time
		select {
		case res = <-resultCh:
			// cancel()
			return res
		case <-time.After(wait):
			// halve wait time, but not below minWait
			if wait > r.ReqTime.MinWait {
				wait = r.ReqTime.NextWaitTime(wait)
				if wait < r.ReqTime.MinWait {
					wait = r.ReqTime.MinWait
				}
			}
			continue
		case <-ctx.Done():
			return Result[R]{Val: zero, Err: errors.New("overall timeout reached")}
		}
	}

	// wait for final results
	select {
	case res = <-resultCh:
		// cancel()
		return res
	case <-ctx.Done():
		return Result[R]{Val: zero, Err: errors.New("overall timeout reached")}
	case <-time.After(wait):
		return Result[R]{Val: zero, Err: errors.New("all requests failed or timed out")}
	}
}

func (r *ReqParam[T, R]) doRequest(ctx context.Context, target T, resultCh chan Result[R], wg *sync.WaitGroup) {
	defer wg.Done()

	ret := r.ReqFunc(ctx, target)
	if ret.Err != nil || !ret.Val.IsValid() {
		return
	}

	// Attempt to send result. Use recover to handle case where channel might be closed.
	// Note: In Go, you cannot directly check if a channel is closed before sending.
	// Sending to a closed channel will panic, so we use recover as a safety measure.
	func() {
		defer func() {
			if recover() != nil {
				// Channel was closed, ignore the panic
			}
		}()
		select {
		case resultCh <- ret:
		case <-ctx.Done():
		}
	}()
}

// weightedPick randomly selects an unused target based on weights
func weightedPick[T Weighted](targets []T, used map[int]bool) int {
	total := 0
	for i, t := range targets {
		if used[i] {
			continue
		}

		w := t.Weight()
		if w < 1 {
			w = 1
		}
		total += w
	}
	if total == 0 {
		return -1
	}

	r := rand.Intn(total)
	for i, t := range targets {
		if used[i] {
			continue
		}

		w := t.Weight()
		if w < 1 {
			w = 1
		}
		if r < w {
			return i
		}

		r -= w
	}

	return -1
}
