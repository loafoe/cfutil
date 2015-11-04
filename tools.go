package cfutil

import (
	"os"
)

func IsFirstInstance() bool {
	appEnv, err := Current()

	if err == nil && appEnv.Index == 0 {
		return true
	}
	return false
}

func ListenString() string {
	return ":" + os.Getenv("PORT")
}

func Getenv(name string) string {
	return os.Getenv(name)
}
