package cfutil

import (
	"os"
)

func ForceHTTP() bool {
	if os.Getenv("FORCE_HTTP") == "true" {
		return true
	}
	return false
}
