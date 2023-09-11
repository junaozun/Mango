package tokenLimit

import (
	"github.com/go-redis/redis"
	"sync"
	"sync/atomic"
	"time"
)

type ITokenLimiterMgr interface {
	GetRedisAlive() *uint32
	StartMonitor()
}

// TokenLimiterMgr 接口限流管理器
type TokenLimiterMgr struct {
	client         *redis.Client
	rescueLock     sync.Mutex
	redisAlive     uint32
	monitorStarted bool
	apiTokenLock   sync.Mutex
	apiTokenLimits map[string]*TokenLimiter // api uniqueKey ---> tokenLimiter
}

func NewTokenLimiterMgr(client *redis.Client) *TokenLimiterMgr {
	return &TokenLimiterMgr{
		client:         client,
		redisAlive:     1,
		monitorStarted: false,
		apiTokenLimits: make(map[string]*TokenLimiter),
	}
}

func (m *TokenLimiterMgr) GetOrCreateTokenLimiter(rate, burst int, uniqueKey string) *TokenLimiter {
	m.apiTokenLock.Lock()
	defer m.apiTokenLock.Unlock()
	v, ok := m.apiTokenLimits[uniqueKey]
	if ok {
		return v
	}
	tl := NewTokenLimiter(rate, burst, uniqueKey, m.client, m)
	m.apiTokenLimits[uniqueKey] = tl
	return tl
}

func (m *TokenLimiterMgr) GetRedisAlive() *uint32 {
	return &m.redisAlive
}

func (m *TokenLimiterMgr) StartMonitor() {
	m.rescueLock.Lock()
	defer m.rescueLock.Unlock()

	if m.monitorStarted {
		return
	}

	m.monitorStarted = true
	atomic.StoreUint32(&m.redisAlive, 0)

	go m.waitForRedis()
}

func (m *TokenLimiterMgr) waitForRedis() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		m.rescueLock.Lock()
		m.monitorStarted = false
		m.rescueLock.Unlock()
	}()

	for range ticker.C {
		if m.Ping() {
			atomic.StoreUint32(&m.redisAlive, 1)
			return
		}
	}
}

func (m *TokenLimiterMgr) Ping() bool {
	v, err := m.client.Ping().Result()
	if err != nil {
		return false
	}
	return v == "PONG"
}
