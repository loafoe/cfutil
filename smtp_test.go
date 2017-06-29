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
	os.Setenv("CF_LOCAL_SMTP", "smtp://foox:tuH5hM/yYtt@email-smtp.us-east-1.amazonaws.com:587")
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
	assert.Equal(t, uri.User.Username(), "foox")
	password, ok := uri.User.Password()
	assert.Equal(t, ok, true)
	assert.Equal(t, password, "tuH5hM/yYtt")
	assert.Equal(t, uri.Hostname(), "email-smtp.us-east-1.amazonaws.com")
	assert.Equal(t, uri.Port(), "587")
}
