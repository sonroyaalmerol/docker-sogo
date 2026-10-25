package sogo

import (
	"database/sql"
	"sync"

	"github.com/bradfitz/gomemcache/memcache"
)

type SOGoConfig struct {
	ACLDbUrl      string `plist:"OCSAclURL"`
	UsersDbUrl    string `plist:"SOGoProfileURL"`
	SessionsDbUrl string `plist:"OCSSessionsFolderURL"`
}

type DBConfig struct {
	Driver string
	DSN    string
	Table  string
}

type SogoService struct {
	aclDB          *sql.DB
	usersDB        *sql.DB
	sessionsDB     *sql.DB
	cache          *memcache.Client
	aclConfig      DBConfig
	usersConfig    DBConfig
	sessionsConfig DBConfig
	mu             sync.Mutex
}
