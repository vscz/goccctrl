package goccctrl

import (
	"context"
	"testing"
	"time"
)

type mockTarget struct {
	id     string
	weight int
	delay  time.Duration
	ok     bool
}

func (m mockTarget) Weight() int { return m.weight }

type mockResult struct {
	val string
	ok  bool
}

func (r mockResult) IsValid() bool { return r.ok }

func mockRequest(ctx context.Context, t mockTarget) Result[mockResult] {
	select {
	case <-ctx.Done():
		return Result[mockResult]{Val: mockResult{"", false}, Err: ctx.Err()}
	case <-time.After(t.delay):
		return Result[mockResult]{Val: mockResult{t.id, t.ok}, Err: nil}
	}
}

func TestProgressiveRequest_1SuccessWithinMaxTime(t *testing.T) {
	targets := []mockTarget{
		{id: "fast", weight: 5, delay: 200 * time.Millisecond, ok: true},
		{id: "slow", weight: 1, delay: 2 * time.Second, ok: true},
	}

	reqTime := ReqTime{TotalTimeout: 3 * time.Second, FirstWait: 1 * time.Second, MinWait: 200 * time.Millisecond}
	param := NewReqParam(targets, reqTime, mockRequest)

	res := param.Do(context.Background())
	if res.Err != nil {
		t.Errorf("expected success, got error: %v", res.Err)
	}
	if res.Val.val != "fast" {
		t.Errorf("expected fast, got %v", res.Val.val)
	}

	t.Logf("res: %+v", res)

	// read more from moreValCh
	for more := range res.MoreValCh {
		t.Logf("more: %+v", more)
	}

	t.Error("test")
}

func TestProgressiveRequest_MultiSuccessWithinMaxTime(t *testing.T) {
	targets := []mockTarget{
		{id: "fast1", weight: 5, delay: 1700 * time.Millisecond, ok: true},
		{id: "fast2", weight: 5, delay: 1200 * time.Millisecond, ok: true},
		{id: "fast3", weight: 5, delay: 1200 * time.Millisecond, ok: true},
		{id: "slow", weight: 1, delay: 3 * time.Second, ok: true},
	}

	reqTime := ReqTime{TotalTimeout: 5 * time.Second, FirstWait: 1 * time.Second, MinWait: 200 * time.Millisecond}
	param := NewReqParam(targets, reqTime, mockRequest)

	res := param.Do(context.Background())
	if res.Err != nil {
		t.Errorf("expected success, got error: %v", res.Err)
	}
	if res.Val.val != "fast1" {
		t.Errorf("expected fast1, got %v", res.Val.val)
	}

	t.Logf("res: %+v", res)

	// read more from moreValCh
	for more := range res.MoreValCh {
		t.Logf("more: %+v", more)
	}

	t.Error("test")
}

func TestProgressiveRequest_SomeTimeout(t *testing.T) {
	targets := []mockTarget{
		{id: "timeout", weight: 1, delay: 3 * time.Second, ok: true},
		{id: "timeout2", weight: 1, delay: 4 * time.Second, ok: true},
		{id: "success", weight: 1, delay: 1 * time.Second, ok: true},
		{id: "success2", weight: 1, delay: 2 * time.Second, ok: true},
	}

	reqTime := ReqTime{2500 * time.Millisecond, 1 * time.Second, 200 * time.Millisecond}
	param := NewReqParam(targets, reqTime, mockRequest)

	res := param.Do(context.Background())
	if res.Err == nil {
		t.Errorf("expected timeout error, got %v", res.Err)
	}
	if res.Val.val != "timeout" {
		t.Errorf("expected timeout, got %v", res.Val.val)
	}

	t.Logf("res: %+v", res)
	// read more from moreValCh
	for more := range res.MoreValCh {
		t.Logf("more: %+v", more)
	}

	t.Error("test")
}

func TestProgressiveRequest_AllTimeout(t *testing.T) {
	targets := []mockTarget{
		{id: "timeout", weight: 1, delay: 3 * time.Second, ok: true},
		{id: "timeout2", weight: 1, delay: 4 * time.Second, ok: true},
	}

	reqTime := ReqTime{500 * time.Millisecond, 1 * time.Second, 200 * time.Millisecond}
	param := NewReqParam(targets, reqTime, mockRequest)

	res := param.Do(context.Background())
	if res.Err == nil {
		t.Errorf("expected timeout error, got %v", res.Err)
	}
	if res.Val.val != "timeout" {
		t.Errorf("expected timeout, got %v", res.Val.val)
	}

	t.Logf("res: %+v", res)
	// read more from moreValCh
	for more := range res.MoreValCh {
		if more.Val.val != "timeout2" {
			t.Errorf("more: expected timeout2, got %v", more.Val.val)
		}

		t.Logf("more: %+v", more)
	}

	t.Error("test")
}
