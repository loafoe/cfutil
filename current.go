package cfutil

import (
	cfenv "github.com/cloudfoundry-community/go-cfenv"
	"os"
)

func Current() (appEnv *cfenv.App, err error) {
	appEnv, err = cfenv.Current()
	if err != nil {
		os.Setenv("VCAP_APPLICATION", localVcapApplication())
		os.Setenv("MEMORY_LIMIT", localMemoryLimit())
		os.Setenv("VCAP_SERVICES", localVcapServices())
		os.Setenv("CF_LOCAL", "true")
		appEnv, err = cfenv.Current()
	}
	return appEnv, err
}
