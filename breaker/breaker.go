package breaker

import (
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

type Counts struct {
	Requests             int64  // 请求数量
	TotalSuccess         uint64 // 总成功数
	TotalFailures        uint32 // 总失败数
	ConsecutiveSuccesses uint32 // 连续成功数
	ConsecutiveFailures  uint32 // 连续失败数量
}

func (c *Counts) OnRequest() {
	c.Requests++
}

func (c *Counts) OnSuccess() {
	c.TotalSuccess++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
}

func (c *Counts) OnFail() {
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
}

func (c *Counts) Clear() {
	c = new(Counts)
}

type CircuitBreaker struct {
	name          string
	maxRequests   uint32
	interval      time.Duration
	timeout       time.Duration
	readyToTrip   func(counts Counts) bool // 是否启用熔断
	isSuccessful  func(err error) bool
	onStateChange func(name string, form State, to State)
	mutex         sync.Mutex
	state         State
	generation    uint64
	counts        Counts
	expiry        time.Time
}

func NewCircuitBreaker(name string) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:        name,
		maxRequests: 1,
		interval:    0,
		timeout:     time.Duration(20) * time.Second,
		readyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
		isSuccessful: func(err error) bool {
			return err == nil
		},
		onStateChange: nil,
		mutex:         sync.Mutex{},
		state:         0,
		generation:    0,
		counts:        Counts{},
		expiry:        time.Time{},
	}
	cb.Generation()
	return cb
}

func (cb *CircuitBreaker) Generation() {
	cb.counts.Clear()
	cb.generation++
	switch cb.state {
	case StateClosed:
		cb.expiry = time.Now().Add(cb.interval)
	case StateHalfOpen:
		cb.expiry = time.Now()
	case StateOpen:
		cb.expiry = time.Now().Add(cb.timeout)
	}
}

func (cb *CircuitBreaker) Execute(req func() (any, error)) (any, error) {
	// 请求之前判断是否执行熔断器
	err := cb.beforeRequest()
	if err != nil {
		return nil, err
	}
	// 执行请求
	result, err := req()
	cb.counts.OnRequest()

	// 请求之后判断当前状态是否需要变更
	err = cb.afterRequest(cb.isSuccessful(err))
	return result, nil
}

func (cb *CircuitBreaker) beforeRequest() error {
	// 判断当前状态，如果断路器打开状态，返回err
}

func (cb *CircuitBreaker) afterRequest(isSuccess bool) error {
	if isSuccess {
		cb.counts.OnSuccess()
	} else {
		cb.counts.OnFail()
	}
}
