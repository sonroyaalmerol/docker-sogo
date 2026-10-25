package utils

import (
	"fmt"
	"strings"
)

// SystemPrincipals are ACL c_uid values that name no real user. SOGo's
// manage-acl resolves them specially and they must never be passed as the
// subscribe target.
var SystemPrincipals = map[string]struct{}{
	"<default>": {},
	"ALL":       {},
	"anonymous": {},
}

// SubscribeTarget converts an ACL c_uid into the identifier manage-acl expects,
// reporting ok=false for system principals. Groups are stored as "@name" but
// manage-acl wants the bare name: it looks the identifier up with
// contactInfosForUserWithUIDorEmail: and expands the members itself.
func SubscribeTarget(cUID string) (target string, ok bool) {
	if _, isSystem := SystemPrincipals[cUID]; isSystem {
		return "", false
	}

	target = strings.TrimPrefix(cUID, "@")
	if target == "" {
		return "", false
	}

	return target, true
}

func ParsePath(input string) (username string, path string, err error) {
	if !strings.HasPrefix(input, "/") {
		err = fmt.Errorf("invalid format: input %q does not start with '/'", input)
		return "", "", err
	}

	parts := strings.SplitN(input, "/", 3)

	if len(parts) < 3 {
		err = fmt.Errorf(
			"invalid format: expected '/username/path', got %q",
			input,
		)
		return "", "", err
	}

	username = parts[1]
	path = parts[2]

	return username, path, nil
}
