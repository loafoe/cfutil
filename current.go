package cfutil

import (
	cfenv "github.com/cloudfoundry-community/go-cfenv"
	"log"
	"os"
)

func Current() (appEnv *cfenv.App, err error) {
	appEnv, err = cfenv.Current()
	if err != nil {
		log.Print("Simulating local CF")
		os.Setenv("VCAP_APPLICATION", LocalVcapApplication())
		os.Setenv("MEMORY_LIMIT", LocalMemoryLimit())
		os.Setenv("VCAP_SERVICES", LocalVcapServices())
		appEnv, err = cfenv.Current()
	}
	return appEnv, err
}
