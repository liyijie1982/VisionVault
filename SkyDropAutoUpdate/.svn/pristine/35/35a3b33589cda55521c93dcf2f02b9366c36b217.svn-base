package utils

import (
	"fmt"
	"io/ioutil"
	"log"
)

func ReadConfig(workspace string) string {

	f, err := ioutil.ReadFile(workspace + "version.ini")
	if err != nil {
		log.Println("read config failed.", err)
		return ""
	}

	return string(f)
}

func WriteConfig(workspace, version string) {
	fileName := workspace + "version.ini"

	err := ioutil.WriteFile(fileName, []byte(version), 0666)
	if err != nil {
		fmt.Println("write config failed", err)
	}
	fmt.Println("write config success")
}
