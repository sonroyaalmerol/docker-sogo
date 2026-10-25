// Package lock provides a memcached mutex so scheduled jobs run once across replicas.
package lock

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"howett.net/plist"
)

const (
	keyPrefix           = "sogo-tool-plus:lock:"
	defaultMemcachePort = "11211"
)

type memcachedConfig struct {
	Host string `plist:"SOGoMemcachedHost"`
}

// Lock is a held memcached mutex; a nil Lock means the caller should proceed uncoordinated.
type Lock struct {
	client *memcache.Client
	key    string
	holder string
}

// MemcachedHosts reads SOGoMemcachedHost from a SOGo plist; empty means not configured.
func MemcachedHosts(configFile string) ([]string, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("could not open config file %q: %w", configFile, err)
	}
	defer f.Close()

	var cfg memcachedConfig
	if err := plist.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("could not decode plist config: %w", err)
	}

	var hosts []string
	for _, raw := range strings.Split(cfg.Host, ",") {
		host := strings.TrimSpace(raw)
		if host == "" {
			continue
		}
		if !strings.HasPrefix(host, "/") && !strings.Contains(host, ":") {
			host += ":" + defaultMemcachePort
		}
		hosts = append(hosts, host)
	}

	return hosts, nil
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
	hosts, err := MemcachedHosts(configFile)
	if err != nil {
		log.Printf("lock: cannot read %s (%v), running without coordination", configFile, err)
		return nil, true
	}
	if len(hosts) == 0 {
		log.Printf("lock: SOGoMemcachedHost is not set, running without coordination")
		return nil, true
	}

	client := memcache.New(hosts...)
	client.Timeout = 2 * time.Second

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
