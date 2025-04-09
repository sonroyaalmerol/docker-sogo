package sogo

import (
	"fmt"
	"log"
	"time"
)

func (s *SogoService) DeleteSessionsByCreation(maxDuration time.Duration) error {
	latestAllowed := time.Now().Add(-maxDuration)

	sqlStmt := s.sessionsConfig.normalize(fmt.Sprintf(
		`DELETE FROM %s WHERE c_creationdate < ?`,
		s.sessionsConfig.Table,
	))
	_, err := s.sessionsDB.Exec(sqlStmt, latestAllowed.Unix())
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
