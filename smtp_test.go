package cfutil

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindSMTPService(t *testing.T) {

	// Setup
	os.Setenv("CF_LOCAL", "true")
	os.Setenv("CF_LOCAL_SMTP", "smtp://foo:bar@some.host:587")
	_, err := Current()
	if !assert.NoError(t, err) {
		return
	}

	uri, err := FindSMTPService("")

	if !assert.NoError(t, err) {
		t.Errorf("VCAP_SERVICES: %s", os.Getenv("VCAP_SERVICES"))
		return
	}
	if !assert.NotNil(t, uri) {
		return
	}
	fmt.Printf("%v", uri)
	assert.Equal(t, uri.User.Username(), "foo")
	password, ok := uri.User.Password()
	assert.Equal(t, ok, true)
	assert.Equal(t, password, "bar")
	assert.Equal(t, uri.Host, "some.host:587")
}
