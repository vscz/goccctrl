package goccctrl

import (
	"context"
	"testing"
	"time"
)

type mockTargetReq struct {
	id     string
	weight int
	delay  time.Duration
	ok     bool
}

func (m mockTargetReq) Weight() int { return m.weight }

func mockTargetReqFunc(ctx context.Context, t mockTargetReq) Result[mockResult] {
	select {
	case <-ctx.Done():
		return Result[mockResult]{Val: mockResult{"", false}, Err: ctx.Err()}
	case <-time.After(t.delay):
		return Result[mockResult]{Val: mockResult{t.id, t.ok}, Err: nil}
	}
}

func TestProgressiveRequestWaitOther_Success(t *testing.T) {
	targets := []mockTargetReq{
		{"fast", 5, 200 * time.Millisecond, true},
		{"slow", 1, 2 * time.Second, true},
	}

	reqTime := ReqTime{1 * time.Second, 5 * time.Second, 200 * time.Millisecond}
	param := NewReqParamWaitOther(targets, reqTime, mockTargetReqFunc)

	res := param.Do(context.Background())
	if res.Err != nil {
		t.Fatalf("expected success, got error: %v", res.Err)
	}
	if res.Val.val != "fast" {
		t.Fatalf("expected fast, got %v", res.Val.val)
	}

	t.Logf("res: %+v", res)

	// read more from moreValCh
	for more := range res.MoreValCh {
		if more.Val.val != "slow" {
			t.Fatalf("expected slow, got %v", more.Val.val)
		}

		t.Logf("more: %+v", more)
	}
}

func TestProgressiveRequestWaitOther_SomeTimeout(t *testing.T) {
	targets := []mockTargetReq{
		{"timeout", 1, 3 * time.Second, true},
		{"timeout2", 1, 4 * time.Second, true},
		{"success", 1, 1 * time.Second, true},
		{"success2", 1, 2 * time.Second, true},
	}

	reqTime := ReqTime{2500 * time.Millisecond, 1 * time.Second, 200 * time.Millisecond}
	param := NewReqParamWaitOther(targets, reqTime, mockTargetReqFunc)

	res := param.Do(context.Background())
	if res.Err == nil {
		t.Fatalf("expected timeout error, got %v", res.Err)
	}
	if res.Val.val != "timeout" {
		t.Fatalf("expected timeout, got %v", res.Val.val)
	}

	t.Logf("res: %+v", res)
	// read more from moreValCh
	for more := range res.MoreValCh {
		if more.Val.val != "timeout2" {
			t.Fatalf("expected timeout2, got %v", more.Val.val)
		}

		t.Logf("more: %+v", more)
	}
}

func TestProgressiveRequestWaitOther_AllTimeout(t *testing.T) {
	targets := []mockTargetReq{
		{"timeout", 1, 3 * time.Second, true},
		{"timeout2", 1, 4 * time.Second, true},
	}

	reqTime := ReqTime{500 * time.Millisecond, 1 * time.Second, 200 * time.Millisecond}
	param := NewReqParamWaitOther(targets, reqTime, mockTargetReqFunc)

	res := param.Do(context.Background())
	if res.Err == nil {
		t.Fatalf("expected timeout error, got %v", res.Err)
	}
	if res.Val.val != "timeout" {
		t.Fatalf("expected timeout, got %v", res.Val.val)
	}

	t.Logf("res: %+v", res)
	// read more from moreValCh
	for more := range res.MoreValCh {
		if more.Val.val != "timeout2" {
			t.Fatalf("expected timeout2, got %v", more.Val.val)
		}

		t.Logf("more: %+v", more)
	}
}
