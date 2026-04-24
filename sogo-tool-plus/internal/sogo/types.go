package sogo

import (
	"database/sql"
	"sync"
)

type SOGoConfig struct {
	ACLDbUrl                string `plist:"OCSAclURL"`
	UsersDbUrl              string `plist:"SOGoProfileURL"`
	SessionsDbUrl           string `plist:"OCSSessionsFolderURL"`
	OCSFolderInfoURL        string `plist:"OCSFolderInfoURL"`
	OCSStoreURL             string `plist:"OCSStoreURL"`
	OCSCacheFolderURL       string `plist:"OCSCacheFolderURL"`
	OCSEMailAlarmsFolderURL string `plist:"OCSEMailAlarmsFolderURL"`
	OCSAdminURL             string `plist:"OCSAdminURL"`
}

type DBConfig struct {
	Driver string
	DSN    string
	Table  string
}

type SogoService struct {
	aclDB            *sql.DB
	usersDB          *sql.DB
	sessionsDB       *sql.DB
	folderInfoDB     *sql.DB
	aclConfig        DBConfig
	usersConfig      DBConfig
	sessionsConfig   DBConfig
	folderInfoConfig DBConfig
	mu               sync.Mutex
}
