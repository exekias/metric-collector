package util

import "os"

// Getenv returns value of given env var
// or default if env var is empty or doesn't exists
func Getenv(name, dflt string) string {
	if os.Getenv(name) != "" {
		return os.Getenv(name)
	}
	return dflt
}
