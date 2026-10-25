package sogo

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"
)

// sslFallbackParam marks a DSN whose sslmode we chose ourselves, so openDB may
// retry without TLS. It is stripped before the DSN reaches lib/pq.
const sslFallbackParam = "x-sogo-ssl-fallback"

// openDB connects and verifies the connection. For PostgreSQL it emulates
// libpq's "prefer": lib/pq refuses that value, so we request TLS and retry once
// in plaintext when the server has TLS turned off.
func openDB(cfg DBConfig, label string) (*sql.DB, error) {
	dsn, fallback := splitSSLFallback(cfg)

	db, err := sql.Open(cfg.Driver, dsn)
	if err == nil {
		err = db.Ping()
	}
	if err == nil {
		return db, nil
	}
	if db != nil {
		db.Close()
	}

	if !fallback || !isTLSUnsupported(err) {
		return nil, fmt.Errorf("failed to connect to %s DB: %w", label, err)
	}

	log.Printf("%s DB does not support TLS, retrying without it", label)

	plain, plainErr := withSSLMode(dsn, "disable")
	if plainErr != nil {
		return nil, fmt.Errorf("failed to connect to %s DB: %w", label, err)
	}

	db, plainErr = sql.Open(cfg.Driver, plain)
	if plainErr == nil {
		plainErr = db.Ping()
	}
	if plainErr != nil {
		if db != nil {
			db.Close()
		}
		return nil, fmt.Errorf("failed to connect to %s DB: %w", label, plainErr)
	}

	return db, nil
}

func splitSSLFallback(cfg DBConfig) (dsn string, fallback bool) {
	if cfg.Driver != "postgres" {
		return cfg.DSN, false
	}

	parsed, err := url.Parse(cfg.DSN)
	if err != nil {
		return cfg.DSN, false
	}

	query := parsed.Query()
	if query.Get(sslFallbackParam) == "" {
		return cfg.DSN, false
	}

	query.Del(sslFallbackParam)
	parsed.RawQuery = query.Encode()

	return parsed.String(), true
}

func withSSLMode(dsn, mode string) (string, error) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}

	query := parsed.Query()
	query.Set("sslmode", mode)
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}

func isTLSUnsupported(err error) bool {
	msg := strings.ToLower(err.Error())

	return strings.Contains(msg, "ssl is not enabled") ||
		strings.Contains(msg, "server does not support ssl") ||
		strings.Contains(msg, "unsupported sslmode")
}
