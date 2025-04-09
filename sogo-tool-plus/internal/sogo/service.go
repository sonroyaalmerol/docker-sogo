package sogo

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"howett.net/plist"
)

func NewSogoService(configFile string) (*SogoService, error) {
	s := &SogoService{}

	log.Printf("Reading SOGo configuration from: %s", configFile)
	config, err := readSogoConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read sogo config: %w", err)
	}

	log.Println("Parsing database connection details...")
	s.aclConfig, err = parseSogoDSN(config.ACLDbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ACL DB URL: %w", err)
	}

	s.usersConfig, err = parseSogoDSN(config.UsersDbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Users DB URL: %w", err)
	}

	s.sessionsConfig, err = parseSogoDSN(config.SessionsDbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Sessions DB URL: %w", err)
	}

	log.Printf(
		"Connecting to ACL database (%s driver, DSN: %s)...",
		s.aclConfig.Driver,
		s.aclConfig.DSN,
	)
	s.aclDB, err = sql.Open(s.aclConfig.Driver, s.aclConfig.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open ACL DB connection: %w", err)
	}
	if err := s.aclDB.Ping(); err != nil {
		s.aclDB.Close()
		return nil, fmt.Errorf("failed to ping ACL DB: %w", err)
	}
	log.Println("ACL database connection successful.")

	log.Printf(
		"Connecting to Users database (%s driver, DSN: %s)...",
		s.usersConfig.Driver,
		s.usersConfig.DSN,
	)
	s.usersDB, err = sql.Open(s.usersConfig.Driver, s.usersConfig.DSN)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to open Users DB connection: %w",
			err,
		)
	}
	if err := s.usersDB.Ping(); err != nil {
		s.usersDB.Close()
		return nil, fmt.Errorf("failed to ping Users DB: %w", err)
	}
	log.Println("Users database connection successful.")

	log.Printf(
		"Connecting to Sessions database (%s driver, DSN: %s)...",
		s.sessionsConfig.Driver,
		s.sessionsConfig.DSN,
	)
	s.sessionsDB, err = sql.Open(s.sessionsConfig.Driver, s.sessionsConfig.DSN)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to open Sessions DB connection: %w",
			err,
		)
	}
	if err := s.sessionsDB.Ping(); err != nil {
		s.sessionsDB.Close()
		return nil, fmt.Errorf("failed to ping Sessions DB: %w", err)
	}
	log.Println("Sessions database connection successful.")

	return s, nil
}

func (s *SogoService) Close() {
	if s.aclDB != nil {
		s.aclDB.Close()
		log.Println("ACL database connection closed.")
	}
	if s.usersDB != nil {
		s.usersDB.Close()
		log.Println("Users database connection closed.")
	}
	if s.sessionsDB != nil {
		s.sessionsDB.Close()
		log.Println("Sessions database connection closed.")
	}
}

func readSogoConfig(filePath string) (*SOGoConfig, error) {
	configFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf(
			"could not open config file '%s': %w",
			filePath,
			err,
		)
	}
	defer configFile.Close()

	var sogoConfig SOGoConfig
	decoder := plist.NewDecoder(configFile)
	err = decoder.Decode(&sogoConfig)
	if err != nil {
		return nil, fmt.Errorf("could not decode plist config: %w", err)
	}

	if sogoConfig.ACLDbUrl == "" || sogoConfig.UsersDbUrl == "" || sogoConfig.SessionsDbUrl == "" {
		return nil, errors.New(
			"OCSAclURL, SOGoProfileURL, or OCSSessionsFolderURL missing in config",
		)
	}

	return &sogoConfig, nil
}

func parseSogoDSN(sogoURL string) (DBConfig, error) {
	var driver string
	var dsn string

	parsedURL, err := url.Parse(sogoURL)
	if err != nil {
		return DBConfig{}, fmt.Errorf(
			"could not parse URL '%s': %w",
			sogoURL,
			err,
		)
	}

	tableName := path.Base(parsedURL.Path)
	if tableName == "." || tableName == "/" || tableName == "" {
		return DBConfig{}, fmt.Errorf(
			"could not extract table name from path '%s'",
			parsedURL.Path,
		)
	}

	switch parsedURL.Scheme {
	case "mysql":
		driver = "mysql"
		// Reconstruct DSN (user:pass@tcp(host:port)/dbname?params)
		if parsedURL.User != nil {
			dsn += parsedURL.User.String() + "@"
		}
		host := parsedURL.Host
		if !strings.Contains(host, "(") {
			host = fmt.Sprintf("tcp(%s)", host)
		}
		dsn += host
		dbName := path.Dir(parsedURL.Path)
		if dbName != "" && dbName != "." {
			dsn += dbName
		}
		query := parsedURL.Query()
		if query.Get("tls") == "" {
			query.Set("tls", "preferred")
		}
		// Ensure parseTime for TIMESTAMP/DATETIME columns
		if query.Get("parseTime") == "" {
			query.Set("parseTime", "true")
		}
		dsn += "?" + query.Encode()

	case "postgresql":
		driver = "postgres"
		// lib/pq accepts the URL format directly
		// Ensure sslmode is set, default to 'prefer'
		query := parsedURL.Query()
		if query.Get("sslmode") == "" {
			query.Set("sslmode", "prefer")
		}
		parsedURL.RawQuery = query.Encode()
		dsn = parsedURL.String()

	default:
		return DBConfig{}, fmt.Errorf(
			"unsupported database scheme '%s' in URL '%s'",
			parsedURL.Scheme,
			sogoURL,
		)
	}

	return DBConfig{Driver: driver, DSN: dsn, Table: tableName}, nil
}

func (s *SogoService) userProfileExists(uid string) (bool, error) {
	var foundUid string
	sqlStmt := fmt.Sprintf(
		`SELECT c_uid FROM %s WHERE c_uid = $1 LIMIT 1`,
		s.usersConfig.Table,
	)
	err := s.usersDB.QueryRow(sqlStmt, uid).Scan(&foundUid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		log.Printf(
			"Error checking user profile existence for '%s': %v",
			uid,
			err,
		)
		return false, fmt.Errorf(
			"database error checking user '%s': %w",
			uid,
			err,
		)
	}
	return true, nil
}

func (s *SogoService) initializeUserProfile(uid string) error {
	sqlStmt := fmt.Sprintf(
		`INSERT INTO %s (c_uid, c_settings) VALUES ($1, $2)`,
		s.usersConfig.Table,
	)
	_, err := s.usersDB.Exec(sqlStmt, uid, `{"Calendar": {}}`)
	if err != nil {
		log.Printf("Error initializing user profile for '%s': %v", uid, err)
		return fmt.Errorf(
			"database error initializing user '%s': %w",
			uid,
			err,
		)
	}
	log.Printf("Initialized profile for user: %s", uid)
	return nil
}

func (s *SogoService) getAllUsers() ([]string, error) {
	sqlStmt := fmt.Sprintf("SELECT c_uid FROM %s", s.usersConfig.Table)
	rows, err := s.usersDB.Query(sqlStmt)
	if err != nil {
		log.Printf("Error querying all users: %v", err)
		return nil, fmt.Errorf("database error getting all users: %w", err)
	}
	defer rows.Close()

	users := make([]string, 0)
	for rows.Next() {
		var cUid string
		if err := rows.Scan(&cUid); err != nil {
			log.Printf("Error scanning user row: %v", err)
			continue
		}
		cUid = strings.TrimSpace(cUid)
		if cUid != "" {
			users = append(users, cUid)
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating user rows: %v", err)
		return nil, fmt.Errorf(
			"database iteration error getting all users: %w",
			err,
		)
	}

	log.Printf("Found %d users", len(users))
	return users, nil
}

func runSogoTool(args ...string) error {
	cmd := exec.Command("sogo-tool", args...)
	log.Printf("Executing: %s", strings.Join(cmd.Args, " "))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error running sogo-tool %v: %s", err, string(output))
		return fmt.Errorf("sogo-tool execution failed: %w", err)
	}
	log.Printf("sogo-tool output: %s", string(output))
	return nil
}
