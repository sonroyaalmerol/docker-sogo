package utils

import (
	"fmt"
	"strings"
)

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
