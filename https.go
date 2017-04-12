package cfutil

import (
	"os"
)

// ForceHTTP() returns true if the environment variable `FORCE_HTTP` contains `true`.
// This is useful for example when you are running your app locally and don't have access
// to working TLS certificates.
func ForceHTTP() bool {
	if os.Getenv("FORCE_HTTP") == "true" {
		return true
	}
	return false
}
