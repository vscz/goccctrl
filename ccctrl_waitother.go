package goccctrl

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

// Weighted target must implement this interface
type TargetReq interface {
	Weighted
}

// NoWeighted is a struct that implements the Weighted interface
type DefaultTargetReq struct {
}

func (n DefaultTargetReq) Weight() int {
	return 1
}

// ReqFunc is the generic request function type
type TargetReqFunc[T TargetReq, R Validatable] func(ctx context.Context, target T) Result[R]

func NewReqParamWaitOther[T TargetReq, R Validatable](targets []T, reqTime ReqTime, reqFunc TargetReqFunc[T, R]) *ReqParamWaitOther[T, R] {
	return &ReqParamWaitOther[T, R]{
		ReqTime: reqTime,
		Targets: targets,
		ReqFunc: reqFunc,
	}
}

type ReqParamWaitOther[T TargetReq, R Validatable] struct {
	ReqTime ReqTime

	Targets        []T
	ReqFunc        TargetReqFunc[T, R]
	ShuffleTargets bool
}

// SetShuffleTargets sets whether to shuffle targets
func (r *ReqParamWaitOther[T, R]) SetShuffleTargets(shuffle bool) {
	r.ShuffleTargets = shuffle
}

// DoShuffleTargets shuffles targets
func (r *ReqParamWaitOther[T, R]) DoShuffleTargets() {
	if !r.ShuffleTargets {
		return
	}

	newTargets := make([]T, 0, len(r.Targets))
	used := make(map[int]bool)

	for len(used) < len(r.Targets) {
		// weighted pick
		idx := targetReqPick(r.Targets, used)
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
func (r *ReqParamWaitOther[T, R]) Do(ctx context.Context) Result[R] {
	var zero R
	if len(r.Targets) == 0 {
		return Result[R]{Val: zero, Err: errors.New("no targets provided")}
	}

	ctx, _ = context.WithTimeout(ctx, r.ReqTime.TotalTimeout)
	// defer cancel()

	resultCh := make(chan Result[R], len(r.Targets)-1)

	wait := r.ReqTime.FirstWait

	r.DoShuffleTargets()
	for idx := range r.Targets {
		go r.doRequest(ctx, r.Targets[idx], resultCh)

		// check next by last result or wait time
		select {
		case res := <-resultCh:
			// cancel()
			res.MoreValCh = resultCh
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
			return Result[R]{Val: zero, Err: errors.New("overall timeout reached"), MoreValCh: resultCh}
		}
	}

	// wait for final results
	select {
	case res := <-resultCh:
		// cancel()
		res.MoreValCh = resultCh
		return res
	case <-ctx.Done():
		return Result[R]{Val: zero, Err: errors.New("overall timeout reached"), MoreValCh: resultCh}
	case <-time.After(wait):
		return Result[R]{Val: zero, Err: errors.New("all requests failed or timed out"), MoreValCh: resultCh}
	}
}

func (r *ReqParamWaitOther[T, R]) doRequest(ctx context.Context, target T, resultCh chan Result[R]) {
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

// targetReqPick randomly selects an unused target based on weights
func targetReqPick[T TargetReq](targets []T, used map[int]bool) int {
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
