package conf

import "fmt"

type conflictKeyError struct {
	key string
}

func newConflictKeyError(key string) conflictKeyError {
	return conflictKeyError{key: key}
}

func (e conflictKeyError) Error() string {
	return fmt.Sprintf("conflict key %s, pay attention to anonymous fields", e.key)
}
