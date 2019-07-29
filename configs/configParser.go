package configs

import (
	"fmt"
	"io/ioutil"
)

func Parse(configFile string) map[string]interface{} {
	dat, err := ioutil.ReadFile("/tmp/dat")
	if err != nil {
		panic(fmt.Sprintf("Error: %v encountered reading %v", err, configFile))
	}
	fmt.Print(string(dat))
	return nil
}
