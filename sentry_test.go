package cfutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSentryDSN(t *testing.T) {

	// Setup
	DSN := "https://foo:bar@some.host/some/path/to/sentry"
	os.Setenv("CF_LOCAL", "true")
	os.Setenv("CF_LOCAL_SENTRY", "sentry:"+DSN)
	_, err := Current()
	if !assert.NoError(t, err) {
		return
	}

	sentryDSN, err := SentryDSN("")

	if !assert.NoError(t, err) {
		t.Errorf("VCAP_SERVICES: %s", os.Getenv("VCAP_SERVICES"))
		return
	}
	assert.Equal(t, sentryDSN, DSN)
}
