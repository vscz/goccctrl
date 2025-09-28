package main

import (
	"context"
	"fmt"
	"time"

	"github.com/vscz/goccctrl"
)

// ---- Example target ----
type PayChannel struct {
	Name      string
	WeightVal int
}

func (p PayChannel) Weight() int { return p.WeightVal }

type PayChannelNoWeight struct {
	goccctrl.NoWeighted
	Name string
}

// ---- Example result ----
type PayResult struct {
	Link string
	Ok   bool
}

func (r PayResult) IsValid() bool { return r.Ok }

// ---- Example request function ----
func payChannelRequest(ctx context.Context, ch PayChannel) goccctrl.Result[PayResult] {
	switch ch.Name {
	case "upi":
		time.Sleep(500 * time.Millisecond)
		return goccctrl.Result[PayResult]{Val: PayResult{"upi-link", true}, Err: nil}
	case "razorpay":
		time.Sleep(1500 * time.Millisecond)
		return goccctrl.Result[PayResult]{Val: PayResult{"razorpay-link", true}, Err: nil}
	case "paytm":
		time.Sleep(2 * time.Second)
		return goccctrl.Result[PayResult]{Val: PayResult{"paytm-link", true}, Err: nil}
	default:
		return goccctrl.Result[PayResult]{Val: PayResult{"", false}, Err: nil}
	}
}

func payChannelNoWeightRequest(ctx context.Context, ch PayChannelNoWeight) goccctrl.Result[PayResult] {
	switch ch.Name {
	case "upi":
		time.Sleep(500 * time.Millisecond)
		return goccctrl.Result[PayResult]{Val: PayResult{"upi-link", true}, Err: nil}
	case "razorpay":
		time.Sleep(1500 * time.Millisecond)
		return goccctrl.Result[PayResult]{Val: PayResult{"razorpay-link", true}, Err: nil}
	case "paytm":
		time.Sleep(2 * time.Second)
		return goccctrl.Result[PayResult]{Val: PayResult{"paytm-link", true}, Err: nil}
	default:
		return goccctrl.Result[PayResult]{Val: PayResult{"", false}, Err: nil}
	}
}

func main() {
	targets := []PayChannel{
		{"paytm", 1},
		{"razorpay", 3},
		{"upi", 5},
	}

	reqTime := goccctrl.ReqTime{
		TotalTimeout: 8 * time.Second,
		FirstWait:    2 * time.Second,
		MinWait:      200 * time.Millisecond,
	}
	param := goccctrl.NewReqParam(targets, reqTime, payChannelRequest)
	param.SetShuffleTargets(true)
	result := param.Do(context.Background())
	if result.Err != nil {
		fmt.Println("Error:", result.Err)
	} else {
		fmt.Println("Result:", result.Val.Link)
	}

	// Output:
	// Result: upi-link

	targetsNoWeight := []PayChannelNoWeight{
		{Name: "paytm"},
		{Name: "razorpay"},
		{Name: "upi"},
	}
	paramNoWeight := goccctrl.NewReqParam(targetsNoWeight, reqTime, payChannelNoWeightRequest)
	// paramNoWeight.SetShuffleTargets(true)
	resultNoWeight := paramNoWeight.Do(context.Background())
	if resultNoWeight.Err != nil {
		fmt.Println("Error:", resultNoWeight.Err)
	} else {
		fmt.Println("Result:", resultNoWeight.Val.Link)
	}

	// Output:
	// Result: upi-link
}
