/*
landers Apps
Author: landers
Github: github.com/landers1037
*/

package main

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	fmt.Println("JJPKG is an application packager for jjapps in linux.")
	// get os system
	sys := runtime.GOOS
	args := os.Args
	var argsMap map[string]string
	var err error

	fmt.Printf("Your System is %s.\n", sys)
	if len(args) >= 4 {
		argsMap, err = parseArgs(args)
	}else {
		argsMap, err = parseJson()
	}

	if err != nil {
		fmt.Println("error, parse args failed.")
		os.Exit(1)
	}
	// start building
	err = makeBuildCMD(argsMap)
	if err != nil {
		fmt.Printf("error, build binary failed. %s\n", err.Error())
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
	createPID(argsMap["name"])
	fmt.Println("Thanks for using jjpkg. Enjoy it!")
}

// 使用默认顺序打包程序
// 顺序 name id version description
func parseArgs(args []string) (map[string]string, error) {
	var name, id, version, description string
	if len(args) <= 2 {
		return map[string]string{}, errors.New("not enough args.")
	}
	// arg1 is main.go
	file := args[1]
	if len(args) >= 3 {
		name = args[2]
	}else {
		name = strings.Split(args[1], ".")[0]
	}
	if len(args) >= 4 {
		id = args[3]
	}else {
		id = ""
	}
	if len(args) >= 5 {
		version = args[4]
	}else {
		version = "1.0"
	}
	if len(args) >= 6 {
		description = args[5]
	}else {
		description = "default description"
	}

	return map[string]string{
		"file": file,
		"name": name,
		"id": id,
		"version": version,
		"description": description,
	}, nil
}

// parse from file jjpkg.json
func parseJson() (map[string]string, error) {
	rawBytes, e := ioutil.ReadFile("jjpkg.json")
	if e !=nil {
		return nil, e
	}

	return map[string]string{
		"file": gjson.GetBytes(rawBytes, "compile_entry").String(),
		"name": gjson.GetBytes(rawBytes, "app_info.name").String(),
		"id": gjson.GetBytes(rawBytes, "app_info.id").String(),
		"version": gjson.GetBytes(rawBytes, "app_info.version").String(),
		"description": gjson.GetBytes(rawBytes, "app_info.description").String(),
	}, nil
}

// build
func makeBuildCMD(argsMap map[string]string) error {
	checkRes := checkGo()
	if !checkRes {
		return errors.New("No Go compiler.")
	}

	cmd := fmt.Sprintf("go build -v -o %s -ldflags=\"-w -s\" -tags %s %s", argsMap["id"], argsMap["version"], argsMap["file"])
	sys := runtime.GOOS
	if sys == "darwin" {
		c, err := exec.Command("zsh", "-c", cmd).Output()
		fmt.Println(string(c))
		return err
	}else if sys == "linux" {
		c, err := exec.Command("bash", "-c", cmd).Output()
		fmt.Println(string(c))
		return err
	}else {
		return errors.New("windows not supported.")
	}
}

// check go compiler
func checkGo() bool {
	o, err := exec.Command("which", "go").Output()
	if err != nil || len(o) <= 0 {
		return false
	}
	return true
}

// generate sha256
func createSHA(appID string) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("sha256sum %s > %s.sha256", appID, appID))
	e := cmd.Run()
	if e != nil {
		fmt.Printf("error, createSHA failed. %s\n", e.Error())
	}
	fmt.Println("success, createSHA.")
}

// create version tag
func createVersionTag(appVersion string) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("echo v%s > version", appVersion))
	e := cmd.Run()
	if e != nil {
		fmt.Printf("error, createVersion failed. %s\n", e.Error())
	}
	fmt.Println("success, createVersion.")
}

// create what's this
func createOwnjj(appName, appId, appVer, appDes string) {
	jsonData := fmt.Sprintf("{\"app_name\": \"%s\", \"app_id\": \"%s\", \"app_version\": \"%s\", \"app_des\": \"%s\"}",
		appName, appId, appVer, appDes)
	e := ioutil.WriteFile("app.jj", []byte(jsonData), 0644)
	if e != nil {
		fmt.Printf("error, createjj failed. %s\n", e.Error())
	}
	fmt.Println("success, createjj file.")
}

// create pid file
func createPID(appName string) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("touch %s.pid", appName))
	e := cmd.Run()
	if e != nil {
		fmt.Printf("error, createPID failed. %s\n", e.Error())
	}
	fmt.Println("success, createPID.")
}