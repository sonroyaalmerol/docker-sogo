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

	if config.ACLDbUrl != "" {
		s.aclConfig, err = parseSogoDSN(config.ACLDbUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ACL DB URL: %w", err)
		}
	} else {
		log.Println("WARNING: OCSAclURL not configured. ACL-related features will be unavailable.")
	}

	s.usersConfig, err = parseSogoDSN(config.UsersDbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Users DB URL: %w", err)
	}

	s.sessionsConfig, err = parseSogoDSN(config.SessionsDbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Sessions DB URL: %w", err)
	}

	if config.ACLDbUrl != "" {
		log.Printf(
			"Connecting to ACL database (%s driver)...",
			s.aclConfig.Driver,
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
	}

	log.Printf(
		"Connecting to Users database (%s driver)...",
		s.usersConfig.Driver,
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
		"Connecting to Sessions database (%s driver)...",
		s.sessionsConfig.Driver,
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

	if sogoConfig.UsersDbUrl == "" || sogoConfig.SessionsDbUrl == "" {
		return nil, errors.New(
			"SOGoProfileURL or OCSSessionsFolderURL missing in config",
		)
	}

	return &sogoConfig, nil
}

func parseSogoDSN(sogoURL string) (DBConfig, error) {
	parsedURL, err := url.Parse(sogoURL)
	if err != nil {
		return DBConfig{}, fmt.Errorf(
			"could not parse URL '%s': %w",
			sogoURL,
			err,
		)
	}

	// SOGo URLs always end with /database/table.
	// The last path segment is the table name.
	tableName := path.Base(parsedURL.Path)
	if tableName == "." || tableName == "/" || tableName == "" {
		return DBConfig{}, fmt.Errorf(
			"could not extract table name from path '%s'",
			parsedURL.Path,
		)
	}

	// The database name is the second-to-last path segment.
	dbPath := path.Dir(parsedURL.Path) // strips /table
	dbName := path.Base(dbPath)
	if dbName == "." || dbName == "/" || dbName == "" {
		return DBConfig{}, fmt.Errorf(
			"could not extract database name from path '%s'",
			parsedURL.Path,
		)
	}

	switch parsedURL.Scheme {
	case "postgresql":
		// Build a lib/pq-compatible URL: postgres://user:pass@host:port/dbname?sslmode=...
		// Must strip the table name from the path so lib/pq sees only /dbname.
		dsnURL := *parsedURL // shallow copy
		dsnURL.Path = dbPath
		dsnURL.RawPath = dbPath
		query := dsnURL.Query()
		if query.Get("sslmode") == "" {
			if sslmode := os.Getenv("PGSSLMODE"); sslmode != "" {
				query.Set("sslmode", sslmode)
			} else {
				query.Set("sslmode", "prefer")
			}
		}
		dsnURL.RawQuery = query.Encode()
		return DBConfig{
			Driver: "postgres",
			DSN:    dsnURL.String(),
			Table:  tableName,
		}, nil

	case "mysql":
		// Build go-sql-driver/mysql DSN: user:pass@tcp(host:port)/dbname?tls=...&parseTime=true
		var dsn string
		if parsedURL.User != nil {
			dsn += parsedURL.User.String() + "@"
		}
		host := parsedURL.Host
		if !strings.Contains(host, "(") {
			host = fmt.Sprintf("tcp(%s)", host)
		}
		dsn += host + dbPath
		query := parsedURL.Query()
		if query.Get("tls") == "" {
			if pgsslmode := os.Getenv("PGSSLMODE"); pgsslmode != "" {
				if pgsslmode == "disable" || pgsslmode == "allow" || pgsslmode == "prefer" {
					query.Set("tls", "false")
				} else {
					query.Set("tls", "preferred")
				}
			} else {
				query.Set("tls", "preferred")
			}
		}
		if query.Get("parseTime") == "" {
			query.Set("parseTime", "true")
		}
		dsn += "?" + query.Encode()
		return DBConfig{
			Driver: "mysql",
			DSN:    dsn,
			Table:  tableName,
		}, nil

	case "oracle":
		// Build Oracle EZConnect DSN: user/pass@host:port/servicename
		var user, pass string
		if parsedURL.User != nil {
			user = parsedURL.User.Username()
			pass, _ = parsedURL.User.Password()
		}
		dsn := fmt.Sprintf("%s/%s@%s/%s", user, pass, parsedURL.Host, dbName)
		return DBConfig{
			Driver: "oracle",
			DSN:    dsn,
			Table:  tableName,
		}, nil

	default:
		return DBConfig{}, fmt.Errorf(
			"unsupported database scheme '%s' in URL '%s'",
			parsedURL.Scheme,
			sogoURL,
		)
	}
}

func (s *SogoService) userProfileExists(uid string) (bool, error) {
	var foundUid string
	sqlStmt := s.usersConfig.normalize(fmt.Sprintf(
		`SELECT c_uid FROM %s WHERE c_uid = ? LIMIT 1`,
		s.usersConfig.Table,
	))
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
	sqlStmt := s.usersConfig.normalize(fmt.Sprintf(
		`INSERT INTO %s (c_uid, c_settings) VALUES (?, ?)`,
		s.usersConfig.Table,
	))
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
