package env

import (
	"fmt"
	"os"
)

type EnvVarNotSetError struct {
	key string
}

func (e *EnvVarNotSetError) Error() string {
	return fmt.Sprintf("environment variable not set: %s", e.key)
}

func Get(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return "", &EnvVarNotSetError{key}
	}

	return val, nil
}
