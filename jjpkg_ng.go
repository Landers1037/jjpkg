/*
landers Apps
Author: landers
Github: github.com/landers1037
*/

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
)

// just for local test
// build tool with cgo and full binary
// with json or config file
func main_ng() {
	// get os system
	sys := runtime.GOOS
	f := os.Args[1]

	fmt.Printf("Your System is %s.\n", sys)
	argsMap := readFile(f)
	if len(argsMap) <= 0 {
		fmt.Println("error, parse args failed.")
		os.Exit(1)
	}
	// start building
	err := makeBuildCMD(argsMap)
	if err != nil {
		fmt.Println("error, build binary failed.")
		os.Exit(2)
	}
	fmt.Println("success, build completed.")
	// make more things
	createSHA(argsMap["id"])
	//
	createVersionTag(argsMap["version"])
	//
	createOwnjj(argsMap["name"], argsMap["id"], argsMap["version"], argsMap["description"])
	//
	fmt.Println("Thanks for using jjpkg. Enjoy it!")
}

// file like [main.go+jjapp+app_test+1.0+just for work]
func readFile(f string) map[string]string {
	// only when file is permitted
	_, e := os.Stat(f)
	if e != nil {
		return map[string]string{}
	}
	fileByte, e := ioutil.ReadFile(f)
	if e != nil {
		return map[string]string{}
	}
	fs := string(fileByte)
	input := strings.Split(fs, "+")
	// gen map
	argsMap := make(map[string]string)
	argsMap["file"] = input[0]
	argsMap["name"] = input[1]
	argsMap["id"] = input[2]
	argsMap["version"] = input[3]
	argsMap["description"] = input[4]

	return argsMap
}
