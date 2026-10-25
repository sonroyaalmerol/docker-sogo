// Package lock provides a memcached mutex so scheduled jobs run once across replicas.
package lock

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/sonroyaalmerol/sogo-tool-plus/internal/memcached"
)

const keyPrefix = "sogo-tool-plus:lock:"

// Lock is a held memcached mutex; a nil Lock means the caller should proceed uncoordinated.
type Lock struct {
	client *memcache.Client
	key    string
	holder string
}

func newHolder() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", os.Getpid())
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return fmt.Sprintf("%s/%d/%s", hostname, os.Getpid(), hex.EncodeToString(buf))
}

// Acquire takes the named lock for ttl, reporting acquired=false only when another
// holder owns it. Missing or unreachable memcached fails open so a cache outage never
// silently stops scheduled work.
func Acquire(configFile, name string, ttl time.Duration) (lock *Lock, acquired bool) {
	client, err := memcached.New(configFile)
	if err != nil {
		log.Printf("lock: cannot read %s (%v), running without coordination", configFile, err)
		return nil, true
	}
	if client == nil {
		log.Printf("lock: SOGoMemcachedHost is not set, running without coordination")
		return nil, true
	}

	holder := newHolder()
	item := &memcache.Item{
		Key:        keyPrefix + name,
		Value:      []byte(holder),
		Expiration: int32(ttl.Seconds()),
	}

	switch err := client.Add(item); {
	case err == nil:
		return &Lock{client: client, key: item.Key, holder: holder}, true
	case errors.Is(err, memcache.ErrNotStored):
		log.Printf("lock: %q is held by another replica, skipping", name)
		return nil, false
	default:
		log.Printf("lock: memcached unavailable (%v), running without coordination", err)
		return nil, true
	}
}

// Release drops the lock only if this process still owns it; a nil Lock is a no-op.
func (l *Lock) Release() {
	if l == nil {
		return
	}

	item, err := l.client.Get(l.key)
	if err != nil || string(item.Value) != l.holder {
		return
	}

	if err := l.client.Delete(l.key); err != nil && !errors.Is(err, memcache.ErrCacheMiss) {
		log.Printf("lock: could not release %q: %v", l.key, err)
	}
}
