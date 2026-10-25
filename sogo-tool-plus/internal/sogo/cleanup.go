package sogo

import (
	"fmt"
	"log"
)

type CleanupReport struct {
	BogusProfiles   []string
	DeletedProfiles int
	SkipPrincipals  []string
}

// CleanupBogusProfiles deletes sogo_user_profile rows keyed by an ACL principal
// instead of a login, left by older cal-subscribe-all. It spares "anonymous",
// which SOGo uses as a real account for public access.
func (s *SogoService) CleanupBogusProfiles(apply bool) (*CleanupReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	report := &CleanupReport{}

	sqlStmt := s.usersConfig.normalize(fmt.Sprintf(
		`SELECT c_uid FROM %s WHERE c_uid LIKE ? OR c_uid = ? OR c_uid = ? OR c_uid = ?`,
		s.usersConfig.Table,
	))

	rows, err := s.usersDB.Query(sqlStmt, "@%", "<default>", "ALL", "anonymous")
	if err != nil {
		return nil, fmt.Errorf("database error scanning user profiles: %w", err)
	}

	var deletable []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			rows.Close()
			return nil, fmt.Errorf("error scanning profile row: %w", err)
		}
		report.BogusProfiles = append(report.BogusProfiles, uid)
		if uid == "anonymous" {
			report.SkipPrincipals = append(report.SkipPrincipals, uid)
			continue
		}
		deletable = append(deletable, uid)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("error iterating profile rows: %w", err)
	}
	rows.Close()

	for _, uid := range report.SkipPrincipals {
		log.Printf(
			"Found profile row for system principal %q; leaving it alone because SOGo uses it for public access",
			uid,
		)
	}

	if len(deletable) == 0 {
		log.Println("No bogus profile rows found.")
		return report, nil
	}

	for _, uid := range deletable {
		if !apply {
			log.Printf("Would delete bogus profile row: %q", uid)
			continue
		}

		delStmt := s.usersConfig.normalize(fmt.Sprintf(
			`DELETE FROM %s WHERE c_uid = ?`,
			s.usersConfig.Table,
		))
		if _, err := s.usersDB.Exec(delStmt, uid); err != nil {
			return report, fmt.Errorf("failed to delete profile %q: %w", uid, err)
		}
		log.Printf("Deleted bogus profile row: %q", uid)
		report.DeletedProfiles++
	}

	return report, nil
}
