package main

import (
	"fmt"
	"io/ioutil"
)

func Parse(configFile string) map[string]interface{} {
	dat, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(fmt.Sprintf("Error: %v encountered reading %v", err, configFile))
	}
	fmt.Print(string(dat))
	return nil
}

func main() {
	fmt.Println("hello")
	Parse("chat.json")
}
