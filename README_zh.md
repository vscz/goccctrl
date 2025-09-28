# goccctrl

[English](README.md) | [中文文档](README_zh.md)

---

## 中文文档

`goccctrl` 是一个 Go 语言的 渐进式请求调度库。
适用于需要依次尝试多个外部服务（如支付通道）的场景，不会一次性并发所有请求，而是逐步发起，直到成功返回。

### 特性
- 渐进式请求调度（等待时间逐步减半，直到最小值）。
- 权重随机排序（动态概率控制，而不是固定顺序）。
- 泛型接口（自定义目标类型和返回结果）。
- 任一成功即返回，其他请求立即丢弃。
- 超时安全，防止 goroutine 泄漏。

### 安装
```bash
go get https://github.com/vscz/goccctrl@v0.1.0
```

### 示例

见 [example/main.go](example/main.go)
