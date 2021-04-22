/*
landers Apps
Author: landers
Github: github.com/landers1037
*/

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	Version = "1.3"
	Build = "2021-04-22"
)

func main() {
	fmt.Println("JJPKG is an application packager for jjapps in linux.")
	// get os system
	sys := runtime.GOOS
	// flags
	upx := flag.Bool("upx", false, "use upx tool")
	plugin := flag.Bool("p", false, "build to plugin")
	mod := flag.String("mod", "", "choose go mod [mod/vendor], if empty use GOPATH")
	analyCode := flag.Bool("a", false, "analy the code")
	force := flag.Bool("f", false, "force rebuild")
	detail := flag.Bool("d", false, "show detail")
	level := flag.String("level", "6", "upx zip level")
	help := flag.Bool("h", false, "show usage")
	flag.Parse()

	if *help {
		fmt.Printf("JJPKG INFO:\nVersion: %s Build: %s\n", Version, Build)
		flag.Usage()
		os.Exit(0)
	}

	if *plugin {
		fmt.Println("build to .so")
		args := flag.Args()
		if len(args) <= 1 {
			fmt.Println("need file input")
		}else {
			fmt.Printf("input file is %s\n", args[1])
			cmd := fmt.Sprintf("go build --buildmode=plugin %s", args[1])
			e := exec.Command("bash", "-c", cmd).Run()
			if e != nil {
				fmt.Printf("build failed %s\n", e.Error())
			}else {
				fmt.Println("done!")
			}
		}
		os.Exit(0)
	}
	args := flag.Args()
	var argsMap map[string]string
	var err error
	var analy bool

	fmt.Printf("Your System is %s.\n", sys)
	fmt.Println("Start to parse compile data.")
	if len(args) >= 4 {
		fmt.Println("Start parse args.")
		argsMap, err = parseArgs(args)
	}else {
		fmt.Println("Start parse jjpkg file.")
		if *analyCode {
			analy = true
		}
		argsMap, err = parseJson()
	}

	if err != nil {
		fmt.Println("error, parse args failed.")
		os.Exit(1)
	}
	fmt.Println("Parse compile data successfully.")
	// start building
	fmt.Println("Start to build app.")
	err = makeBuildCMD(argsMap, analy, *force, *mod, *upx, *level, *detail)
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
	_, e := os.Stat("jjpkg.json")
	if e != nil {
		fmt.Println("There is no jjpkg.json here.")
		return nil, e
	}
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
func makeBuildCMD(argsMap map[string]string, analy, force bool, mod string, upx bool, level string, detail bool) error {
	checkRes := checkGo()
	if !checkRes {
		return errors.New("No Go compiler.")
	}

	var rawCmd string
	var analyString string
	var forceString string

	if analy {
		analyString = "-gcflags=\"-m\""
	}else {
		analyString = ""
	}

	if force {
		forceString = "-a"
	}else {
		forceString = ""
	}

	switch mod {
	case "":
		rawCmd = "export GO111MODULE=off && go build %s -x -o %s -ldflags=\"-w -s\" %s -trimpath -p 2 -tags %s %s"
	case "mod":
		rawCmd = "export GO111MODULE=on && go build -mod=mod %s -x -o %s -ldflags=\"-w -s\" %s -trimpath -p 2 -tags %s %s"
	case "vendor":
		rawCmd = "export GO111MODULE=on && go build -mod=vendor %s -x -o %s -ldflags=\"-w -s\" %s -trimpath -p 2 -tags %s %s"
	default:
		rawCmd = "export GO111MODULE=off && go build %s -x -o %s -ldflags=\"-w -s\" %s -trimpath -p 2 -tags %s %s"
	}

	cmd := fmt.Sprintf(rawCmd, forceString, argsMap["id"], analyString, argsMap["version"], argsMap["file"])
	fmt.Println("Build CMD is ", cmd)
	sys := runtime.GOOS
	if sys == "darwin" {
		c := exec.Command("zsh", "-c", cmd)
		stdout, e := c.StdoutPipe()
		c.Stderr = os.Stdout

		if e != nil {
			fmt.Println("Init cmd pipe failed.")
			return e
		}

		if err := c.Start();err != nil {
			fmt.Println("Cmd start failed.")
			return err
		}

		for {
			tmp := make([]byte, 1024)
			_, err := stdout.Read(tmp)
			if detail {
				fmt.Print(string(tmp))
			}else {
				for _, r := range "-\\|/" {
					fmt.Printf("\r%c", r)
					time.Sleep(400 * time.Millisecond)
				}
			}
			if err != nil {
				break
			}
		}
		if err := c.Wait();err != nil {
			fmt.Println("Failed to compile.")
			return err
		}

		fmt.Println("Compiled.")
		if upx {
			fmt.Println("-upx is specified")
			fmt.Println("upx zip level is " + level)
			fmt.Println(fmt.Sprintf("upx output is: %s-upx",  argsMap["id"]))
			upxString := fmt.Sprintf("upx -%s -o %s-upx %s", level, argsMap["id"], argsMap["id"])
			c := exec.Command("zsh", "-c", upxString)
			stdout, e := c.StdoutPipe()
			c.Stderr = os.Stdout

			if e != nil {
				fmt.Println("Init cmd pipe failed.")
				return e
			}

			if err := c.Start();err != nil {
				fmt.Println("Cmd start failed.")
				return err
			}

			for {
				tmp := make([]byte, 1024)
				_, err := stdout.Read(tmp)
				if detail {
					fmt.Print(string(tmp))
				}else {
					for _, r := range "-\\|/" {
						fmt.Printf("\r%c", r)
						time.Sleep(400 * time.Millisecond)
					}
				}
				if err != nil {
					break
				}
			}
			if err := c.Wait();err != nil {
				fmt.Println("Failed to compile.")
				return err
			}
			fmt.Println("Done with upx.")
			return nil
		}else {
			fmt.Println("-upx is not specified.")
			return nil
		}

	}else if sys == "linux" {
		c := exec.Command("bash", "-c", cmd)

		if detail {
			fmt.Println("Print build log to Terminal.")
			stdout, e := c.StdoutPipe()
			// 这里保证了出现错误直接显示在前台 如果不做重定向则不会显示
			c.Stderr = os.Stdout
			if e != nil {
				fmt.Println("Init cmd pipe failed.")
				return e
			}
			if err := c.Start();err != nil {
				fmt.Println("Cmd start failed.")
				return err
			}
			for {
				tmp := make([]byte, 4)
				_, err := stdout.Read(tmp)
				fmt.Printf("%s", tmp)
				if err != nil {
					_ = stdout.Close()
					break
				}
			}
			if err := c.Wait();err != nil {
				fmt.Println("Failed to compile.")
				return err
			}
		}else {
			fmt.Println("Build progress start.")
			stdout, e := c.StdoutPipe()
			//c.Stderr = os.Stdout
			if e != nil {
				fmt.Println("Init cmd pipe failed.")
				return e
			}

			if err := c.Start();err != nil {
				fmt.Println("Cmd start failed.")
				return err
			}
			for {
				o := make([]byte, 1024)
				_, e := stdout.Read(o)
				fmt.Println(string(o))
				for _, str := range "-\\|/" {
					fmt.Printf("\r%c", str)
					time.Sleep(200 * time.Millisecond)
				}
				if e != nil{
					break
				}
			}
			if err := c.Wait();err != nil {
				fmt.Println("Failed to compile.")
				return err
			}
		}

		fmt.Println("Compiled.")
		if upx {
			fmt.Println("-upx is specified")
			fmt.Println("upx zip level is " + level)
			fmt.Println(fmt.Sprintf("upx output is: %s-upx",  argsMap["id"]))
			upxString := fmt.Sprintf("upx -%s -o %s-upx %s", level, argsMap["id"], argsMap["id"])
			c := exec.Command("bash", "-c", upxString)
			stdout, e := c.StdoutPipe()
			c.Stderr = os.Stdout

			if e != nil {
				fmt.Println("Init cmd pipe failed.")
				return e
			}

			if err := c.Start();err != nil {
				fmt.Println("Cmd start failed.")
				return err
			}

			for {
				// if print detail
				tmp := make([]byte, 1024)
				_, err := stdout.Read(tmp)
				if detail {
					fmt.Print(string(tmp))
				}else {
					for _, r := range "-\\|/" {
						fmt.Printf("\r%c", r)
						time.Sleep(200 * time.Millisecond)
					}
				}

				if err != nil {
					break
				}
			}
			if err := c.Wait();err != nil {
				fmt.Println("Failed to compile.")
				return err
			}
			fmt.Println("Done with upx.")
			return nil
		}else {
			fmt.Println("-upx is not specified.")
			return nil
		}

	}else if sys == "windows" {
		fmt.Println("windows系统需要保证你的go在path路径下")
		in := bytes.NewBuffer(nil)
		var out bytes.Buffer
		c := exec.Command("cmd")
		c.Stdin = in
		c.Stdout = &out
		in.WriteString(cmd + "\n")
		e := c.Start()
		if e != nil {
			return e
		}
		e = c.Wait()
		if e != nil {
			return e
		}

		fmt.Println(out.String())
		return nil
	} else {
		return errors.New("your system is not supported.")
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