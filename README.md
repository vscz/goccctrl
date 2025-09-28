# goccctrl

[English](README.md) | [中文文档](README_zh.md)

---

## English Documentation

`goccctrl` is a Go library for **progressive and probabilistic request scheduling**.  
It is useful when you want to request multiple external services (e.g., payment gateways) but **not all at once** — instead, requests are sent gradually with dynamic wait times and probabilistic ordering.

### Features
- Progressive request scheduling (first wait, then decay with a minimum bound).
- Weighted target selection (probabilistic order, not fixed).
- Generic interface design (define your own `Target` and `Result`).
- Early return when any request succeeds.
- Timeout-safe and goroutine-leak-safe.

### Installation
```bash
go get github.com/vscz/goccctrl@v0.1.0
```

### Example

See [example/main.go](example/main.go)