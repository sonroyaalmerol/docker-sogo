package sogo

import (
	"fmt"
	"log"

	"github.com/sonroyaalmerol/sogo-tool-plus/internal/utils"
)

func (s *SogoService) CalSubscribeUser(uid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Starting subscription process for user: %s", uid)

	exists, err := s.userProfileExists(uid)
	if err != nil {
		return fmt.Errorf("failed to check user profile for '%s': %w", uid, err)
	}
	if !exists {
		log.Printf("User profile for '%s' not found, initializing...", uid)
		if err := s.initializeUserProfile(uid); err != nil {
			return fmt.Errorf(
				"failed to initialize user profile for '%s': %w",
				uid,
				err,
			)
		}
	} else {
		log.Printf("User profile for '%s' already exists.", uid)
	}

	sqlStmt := s.aclConfig.normalize(fmt.Sprintf(`
        SELECT DISTINCT c_object
        FROM %s
        WHERE c_object LIKE ? AND (c_uid = ? OR c_uid = ?) AND c_role <> ?
    `, s.aclConfig.Table))

	rows, err := s.aclDB.Query(
		sqlStmt,
		"%Calendar%",
		uid,
		"<default>",
		"None",
	)
	if err != nil {
		log.Printf("Error querying ACLs for user '%s': %v", uid, err)
		return fmt.Errorf(
			"database error querying ACLs for user '%s': %w",
			uid,
			err,
		)
	}
	defer rows.Close()

	subscribedCount := 0
	for rows.Next() {
		var cObject string
		if err := rows.Scan(&cObject); err != nil {
			log.Printf("Error scanning ACL row for user '%s': %v", uid, err)
			continue
		}

		owner, parsedPath, err := utils.ParsePath(cObject)
		if err != nil {
			log.Printf(
				"Failed to parse calendar path: %v",
				err,
			)
			continue
		}

		if owner == uid {
			log.Printf(
				"Skipping self-subscription for user '%s' to calendar '%s'",
				uid,
				parsedPath,
			)
			continue
		}

		err = runSogoTool(
			"manage-acl",
			"subscribe",
			owner,
			parsedPath,
			uid,
		)
		if err != nil {
			log.Printf(
				"Failed to subscribe user '%s' to calendar '%s': %v",
				uid,
				parsedPath,
				err,
			)
		} else {
			subscribedCount++
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating ACL rows for user '%s': %v", uid, err)
		return fmt.Errorf(
			"database iteration error processing ACLs for user '%s': %w",
			uid,
			err,
		)
	}

	log.Printf(
		"Subscription process for user '%s' completed. Subscribed to %d calendars.",
		uid,
		subscribedCount,
	)
	return nil
}

type subscription struct {
	owner        string
	parsedPath   string
	user         string
	calendarPath string
}

func (s *SogoService) CalSubscribeAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println("Starting subscription process for all users...")

	sqlStmt := s.aclConfig.normalize(fmt.Sprintf(`
        SELECT c_object, c_uid
        FROM %s
        WHERE c_object LIKE ? AND c_role <> ?
    `, s.aclConfig.Table))
	rows, err := s.aclDB.Query(sqlStmt, "%Calendar%", "None")
	if err != nil {
		log.Printf("Error querying all relevant ACLs: %v", err)
		return fmt.Errorf("database error querying all ACLs: %w", err)
	}
	defer rows.Close()

	var subscriptions []subscription

	for rows.Next() {
		var cObject, cUid string
		if err := rows.Scan(&cObject, &cUid); err != nil {
			log.Printf(
				"Error scanning ACL row during cal-subscribe-all: %v",
				err,
			)
			continue
		}

		if cObject == "" {
			log.Printf(
				"Skipping invalid calendar path: %s",
				cObject,
			)
			continue
		}

		owner, parsedPath, err := utils.ParsePath(cObject)
		if err != nil {
			log.Printf(
				"Failed to parse calendar path: %v",
				err,
			)
			continue
		}

		if owner == cUid {
			log.Printf(
				"Skipping self-subscription for user '%s' to calendar '%s'",
				cUid,
				parsedPath,
			)
			continue
		}

		subscriptions = append(subscriptions, subscription{
			owner:        owner,
			parsedPath:   parsedPath,
			user:         cUid,
			calendarPath: cObject,
		})
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating ACL rows during cal-subscribe-all: %v", err)
		return fmt.Errorf(
			"database iteration error processing all ACLs: %w",
			err,
		)
	}

	if len(subscriptions) == 0 {
		log.Println("No subscriptions found to process.")
		return nil
	}

	totalSubscriptions := 0
	totalFailures := 0
	for _, sub := range subscriptions {
		err = runSogoTool(
			"manage-acl",
			"subscribe",
			sub.owner,
			sub.parsedPath,
			sub.user,
		)
		if err != nil {
			log.Printf(
				"Failed to subscribe user '%s' to calendar '%s': %v",
				sub.user,
				sub.calendarPath,
				err,
			)
			totalFailures++
		} else {
			totalSubscriptions++
		}
	}

	log.Printf(
		"Subscription process for all users completed. Performed %d subscriptions with %d failures.",
		totalSubscriptions,
		totalFailures,
	)
	if totalFailures > 0 {
		return fmt.Errorf(
			"encountered %d failures during cal-subscribe-all",
			totalFailures,
		)
	}
	return nil
}
