package H

import (
	"crypto/md5"
	"encoding/hex"
	"sync"
	"time"
)

var DefaultCacheExpiration = 15 * time.Minute
var cleanupJobEvery = 6 * time.Hour

type valueCacheLTE struct {
	value interface{}
	lte   time.Time
}

type UserCache struct {
	cache      map[string]valueCacheLTE
	mu         sync.Mutex
	lastAccess time.Time
}

type CacheSessions struct {
	mu    sync.Mutex
	users map[string]*UserCache
}

var cache CacheSessions

func runCleanupUser(cacheSession *UserCache) *UserCache {
	cacheSession.mu.Lock()
	defer cacheSession.mu.Unlock()
	for key := range cacheSession.cache {
		if cacheSession.cache[key].lte.Before(time.Now()) {
			delete(cacheSession.cache, key)
		}
	}
	return cacheSession
}

func runCleanup() {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	for user_uuid, cacheSession := range cache.users {
		if time.Since(cacheSession.lastAccess) > cleanupJobEvery {
			delete(cache.users, user_uuid)
		}
	}
	ParallelMapWorker(&cache.users, 4, runCleanupUser)
}

func cacheCleanup() {
	for {
		time.Sleep(time.Minute * 10)
		runCleanup()
	}
}

func MD5(text string) string {
	h := md5.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}

func GetCacheSession(user_uuid string) *UserCache {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if sessionCache, ok := cache.users[user_uuid]; ok {
		sessionCache.lastAccess = time.Now()
		return sessionCache
	}
	sessionCache := UserCache{
		cache:      make(map[string]valueCacheLTE),
		lastAccess: time.Now(),
	}
	cache.users[user_uuid] = &sessionCache
	return &sessionCache
}

func ClearUserCacheSession(user_uuid string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	delete(cache.users, user_uuid)
}

func (sessionCache *UserCache) Set(key string, value interface{}, duration time.Duration) {
	sessionCache.mu.Lock()
	defer sessionCache.mu.Unlock()
	sessionCache.cache[key] = valueCacheLTE{
		value: value,
		lte:   time.Now().Add(duration),
	}
}

func (sessionCache *UserCache) Get(key string) interface{} {
	sessionCache.mu.Lock()
	defer sessionCache.mu.Unlock()
	data, ok := sessionCache.cache[key]
	if !ok {
		return nil
	}
	if data.lte.Before(time.Now()) {
		delete(sessionCache.cache, key)
		return nil
	}
	return data.value
}

func (sessionCache *UserCache) Delete(key string) {
	sessionCache.mu.Lock()
	defer sessionCache.mu.Unlock()
	delete(sessionCache.cache, key)
}

func (sessionCache *UserCache) Clear() {
	sessionCache.mu.Lock()
	defer sessionCache.mu.Unlock()
	sessionCache.cache = make(map[string]valueCacheLTE)
}

func (sessionCache *UserCache) Exists(key string) bool {
	return sessionCache.Get(key) != nil
}

func init() {
	cache = CacheSessions{
		users: make(map[string]*UserCache),
	}
	go cacheCleanup()
}
