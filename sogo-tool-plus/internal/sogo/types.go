package sogo

import (
	"database/sql"
	"sync"
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
	aclConfig      DBConfig
	usersConfig    DBConfig
	sessionsConfig DBConfig
	mu             sync.Mutex
}
