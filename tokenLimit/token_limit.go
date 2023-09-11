package tokenLimit

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	xrate "golang.org/x/time/rate"
	"log"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	tokenFormat     = "{%s}.tokens"
	timestampFormat = "{%s}.ts"
	pingInterval    = time.Millisecond * 100
)

// A TokenLimiter controls how frequently events are allowed to happen with in one second.
type TokenLimiter struct {
	rate          int // 速率
	burst         int // 容量
	tokenKey      string
	timestampKey  string
	rescueLimiter *xrate.Limiter // 保底限额
	execScript    *redis.Script  // 执行脚本
	client        *redis.Client
	iMgr          ITokenLimiterMgr
}

// NewTokenLimiter returns a new TokenLimiter that allows events up to rate and permits
// bursts of at most burst tokens.
func NewTokenLimiter(rate, burst int, key string, client *redis.Client, iMgr ITokenLimiterMgr) *TokenLimiter {
	return &TokenLimiter{
		rate:          rate,
		burst:         burst,
		execScript:    redis.NewScript(limitRedisScript),
		client:        client,
		tokenKey:      fmt.Sprintf(tokenFormat, key),
		timestampKey:  fmt.Sprintf(timestampFormat, key),
		rescueLimiter: xrate.NewLimiter(xrate.Every(time.Second/time.Duration(rate)), burst),
		iMgr:          iMgr,
	}
}

// Allow is shorthand for AllowN(time.Now(), 1).
func (t *TokenLimiter) Allow() bool {
	return t.AllowN(time.Now(), 1)
}

func (t *TokenLimiter) AllowN(now time.Time, n int) bool {
	return t.reserveN(now, n)
}

func (t *TokenLimiter) GetRedisAliveFlag() *uint32 {
	return t.iMgr.GetRedisAlive()
}

func (t *TokenLimiter) reserveN(now time.Time, n int) bool {
	if atomic.LoadUint32(t.GetRedisAliveFlag()) == 0 {
		return t.rescueLimiter.AllowN(now, n)
	}

	resp, err := t.execScript.Run(t.client,
		[]string{
			t.tokenKey,
			t.timestampKey,
		},
		[]string{
			strconv.Itoa(t.rate),
			strconv.Itoa(t.burst),
			strconv.FormatInt(now.Unix(), 10),
			strconv.Itoa(n),
		}).Result()
	// redis allowed == false
	// Lua boolean false -> r Nil bulk reply
	if err == redis.Nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		log.Printf("fail to use rate limiter: %s", err)
		return false
	}
	if err != nil {
		log.Printf("fail to use rate limiter: %s, use in-process limiter for rescue", err)
		t.iMgr.StartMonitor()
		return t.rescueLimiter.AllowN(now, n)
	}

	code, ok := resp.(int64)
	if !ok {
		log.Printf("fail to eval redis script: %v, use in-process limiter for rescue", resp)
		t.iMgr.StartMonitor()
		return t.rescueLimiter.AllowN(now, n)
	}

	// redis allowed == true
	// Lua boolean true -> r integer reply with value of 1
	return code == 1
}
