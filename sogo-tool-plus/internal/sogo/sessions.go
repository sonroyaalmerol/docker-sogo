package sogo

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

// DeleteSessionsByCreation drops sessions created before now-maxDuration. It
// mirrors SOGo's own -[SOGoSession deleteValueForSessionKey:], which removes
// the row *and* the "session:<id>" memcached entry; SOGo reads the cache first,
// so a row-only delete would leave the session usable until the cache expires.
func (s *SogoService) DeleteSessionsByCreation(maxDuration time.Duration) error {
	latestAllowed := time.Now().Add(-maxDuration)

	ids, err := s.sessionIDsBefore(latestAllowed)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		log.Printf("No sessions to remove before: %v", latestAllowed)
		return nil
	}

	sqlStmt := s.sessionsConfig.normalize(fmt.Sprintf(
		`DELETE FROM %s WHERE c_creationdate < ?`,
		s.sessionsConfig.Table,
	))
	if _, err := s.sessionsDB.Exec(sqlStmt, latestAllowed.Unix()); err != nil {
		log.Printf("Error removing sessions before '%v': %v", latestAllowed, err)
		return fmt.Errorf(
			"database error removing sessions before '%v': %w",
			latestAllowed,
			err,
		)
	}

	evicted, err := s.evictSessions(ids)
	if err != nil {
		return err
	}

	log.Printf(
		"Removed %d sessions before %v (%d cache entries evicted)",
		len(ids),
		latestAllowed,
		evicted,
	)
	return nil
}

func (s *SogoService) sessionIDsBefore(cutoff time.Time) ([]string, error) {
	sqlStmt := s.sessionsConfig.normalize(fmt.Sprintf(
		`SELECT c_id FROM %s WHERE c_creationdate < ?`,
		s.sessionsConfig.Table,
	))

	rows, err := s.sessionsDB.Query(sqlStmt, cutoff.Unix())
	if err != nil {
		log.Printf("Error listing sessions before '%v': %v", cutoff, err)
		return nil, fmt.Errorf(
			"database error listing sessions before '%v': %w",
			cutoff,
			err,
		)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Printf("Error scanning session row: %v", err)
			return nil, fmt.Errorf("error scanning session row: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating session rows: %v", err)
		return nil, fmt.Errorf("error iterating session rows: %w", err)
	}

	return ids, nil
}

func (s *SogoService) evictSessions(ids []string) (int, error) {
	if s.cache == nil {
		log.Printf(
			"WARNING: memcached not configured; %d expired sessions stay cached until SOGoCacheCleanupInterval elapses",
			len(ids),
		)
		return 0, nil
	}

	evicted := 0
	for _, id := range ids {
		err := s.cache.Delete("session:" + id)
		switch {
		case err == nil:
			evicted++
		case errors.Is(err, memcache.ErrCacheMiss):
		default:
			log.Printf("Failed to evict cached session '%s': %v", id, err)
			return evicted, fmt.Errorf("failed to evict cached session %q: %w", id, err)
		}
	}

	return evicted, nil
}
