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

	sqlStmt := fmt.Sprintf(`
        SELECT DISTINCT c_object
        FROM %s
        WHERE c_object LIKE $1 AND (c_uid = $2 OR c_uid = $3) AND c_role <> $4
    `, s.aclConfig.Table)

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

func (s *SogoService) CalSubscribeAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println("Starting subscription process for all users...")

	users, err := s.getAllUsers()
	if err != nil {
		return fmt.Errorf("failed to get all users for subscription: %w", err)
	}
	if len(users) == 0 {
		log.Println("No users found to subscribe.")
		return nil
	}

	sqlStmt := fmt.Sprintf(`
        SELECT c_object, c_uid
        FROM %s
        WHERE c_object LIKE $1 AND c_role <> $2
    `, s.aclConfig.Table)
	rows, err := s.aclDB.Query(sqlStmt, "%Calendar%", "None")
	if err != nil {
		log.Printf("Error querying all relevant ACLs: %v", err)
		return fmt.Errorf("database error querying all ACLs: %w", err)
	}
	defer rows.Close()

	subscriptionsNeeded := make(map[string]map[string]bool)

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
				"Skipping invalid calendar path derived from: %s",
				cObject,
			)
			continue
		}

		if _, ok := subscriptionsNeeded[cObject]; !ok {
			subscriptionsNeeded[cObject] = make(map[string]bool)
		}

		if cUid == "<default>" {
			for _, user := range users {
				subscriptionsNeeded[cObject][user] = true
			}
		} else {
			userExists := false
			for _, u := range users {
				if u == cUid {
					userExists = true
					break
				}
			}
			if userExists {
				subscriptionsNeeded[cObject][cUid] = true
			} else {
				log.Printf(
					"ACL found for user '%s' but user not in users table, skipping subscription for calendar '%s'",
					cUid,
					cObject,
				)
			}
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating ACL rows during cal-subscribe-all: %v", err)
		return fmt.Errorf(
			"database iteration error processing all ACLs: %w",
			err,
		)
	}

	totalSubscriptions := 0
	totalFailures := 0
	for calendarPath, usersToCalSubscribe := range subscriptionsNeeded {
		for user := range usersToCalSubscribe {
			owner, parsedPath, err := utils.ParsePath(calendarPath)
			if err != nil {
				log.Printf(
					"Failed to parse calendar path: %v",
					err,
				)
				totalFailures++
				continue
			}

			err = runSogoTool(
				"manage-acl",
				"subscribe",
				owner,
				parsedPath,
				user,
			)
			if err != nil {
				log.Printf(
					"Failed to subscribe user '%s' to calendar '%s': %v",
					user,
					calendarPath,
					err,
				)
				totalFailures++
			} else {
				totalSubscriptions++
			}
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
