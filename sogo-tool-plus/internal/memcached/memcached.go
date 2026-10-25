// Package memcached resolves SOGoMemcachedHost and builds a shared client.
package memcached

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"howett.net/plist"
)

const defaultPort = "11211"

type config struct {
	Host string `plist:"SOGoMemcachedHost"`
}

// Hosts reads SOGoMemcachedHost from a SOGo plist; empty means not configured.
func Hosts(configFile string) ([]string, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("could not open config file %q: %w", configFile, err)
	}
	defer f.Close()

	var cfg config
	if err := plist.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("could not decode plist config: %w", err)
	}

	var hosts []string
	for raw := range strings.SplitSeq(cfg.Host, ",") {
		host := strings.TrimSpace(raw)
		if host == "" {
			continue
		}
		if !strings.HasPrefix(host, "/") && !strings.Contains(host, ":") {
			host += ":" + defaultPort
		}
		hosts = append(hosts, host)
	}

	return hosts, nil
}

// New returns a client, or nil when SOGoMemcachedHost is not configured.
func New(configFile string) (*memcache.Client, error) {
	hosts, err := Hosts(configFile)
	if err != nil {
		return nil, err
	}
	if len(hosts) == 0 {
		return nil, nil
	}

	client := memcache.New(hosts...)
	client.Timeout = 2 * time.Second

	return client, nil
}
