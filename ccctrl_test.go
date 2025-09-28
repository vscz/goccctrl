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

func TestProgressiveRequest_Success(t *testing.T) {
	targets := []mockTarget{
		{"fast", 5, 200 * time.Millisecond, true},
		{"slow", 1, 2 * time.Second, true},
	}

	reqTime := ReqTime{1 * time.Second, 5 * time.Second, 200 * time.Millisecond}
	param := NewReqParam(targets, reqTime, mockRequest)

	res := param.Do(context.Background())
	if res.Err != nil {
		t.Fatalf("expected success, got error: %v", res.Err)
	}
	if res.Val.val != "fast" {
		t.Fatalf("expected fast, got %v", res.Val.val)
	}
}

func TestProgressiveRequest_Timeout(t *testing.T) {
	targets := []mockTarget{
		{"timeout", 1, 3 * time.Second, true},
	}

	reqTime := ReqTime{500 * time.Millisecond, 1 * time.Second, 200 * time.Millisecond}
	param := NewReqParam(targets, reqTime, mockRequest)

	res := param.Do(context.Background())
	if res.Err == nil {
		t.Fatalf("expected timeout error, got %v", res.Err)
	}
}
