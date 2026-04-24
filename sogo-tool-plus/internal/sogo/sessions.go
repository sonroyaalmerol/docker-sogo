package sogo

import (
	"fmt"
	"log"
	"time"
)

func (s *SogoService) DeleteSessionsByCreation(maxDuration time.Duration) error {
	latestAllowed := time.Now().Add(-maxDuration)

	// Check that c_creationdate column exists before attempting DELETE
	var checkQuery string
	var checkArgs []any

	switch s.sessionsConfig.Driver {
	case "postgres":
		checkQuery = `SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = 'public' AND table_name = $1 AND column_name = 'c_creationdate'`
		checkArgs = []any{s.sessionsConfig.Table}
	case "mysql":
		checkQuery = `SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = 'c_creationdate'`
		checkArgs = []any{s.sessionsConfig.Table}
	default:
		return fmt.Errorf("unsupported database driver for session expiry check")
	}

	var colCount int
	err := s.sessionsDB.QueryRow(checkQuery, checkArgs...).Scan(&colCount)
	if err != nil {
		log.Printf("Error checking for c_creationdate column: %v", err)
		return fmt.Errorf("failed to check session table schema: %w", err)
	}
	if colCount == 0 {
		return fmt.Errorf(
			"column c_creationdate not found in session table — this feature requires SOGo v5.x with the c_creationdate column in the sessions table",
		)
	}

	sqlStmt := s.sessionsConfig.normalize(fmt.Sprintf(
		`DELETE FROM %s WHERE c_creationdate < ?`,
		s.sessionsConfig.Table,
	))
	_, err = s.sessionsDB.Exec(sqlStmt, latestAllowed.Unix())
	if err != nil {
		log.Printf("Error removing sessions before '%v': %v", latestAllowed, err)
		return fmt.Errorf(
			"database error removing sessions before '%v': %w",
			latestAllowed,
			err,
		)
	}
	log.Printf("Removed sessions before: %v", latestAllowed)
	return nil
}
