package main

import (
	"fmt"
	"log"
	"os"
)

var (
	Version   = ""
	BuildTime = ""
)

func main() {
	if len(os.Args) < 4 {
		if len(os.Args) == 2 && os.Args[1] == "-v" {
			fmt.Printf("version: %s\nbuildTime: %s\n",
				Version, BuildTime)
		} else {
			help()
		}
		return
	}
	envName := os.Args[1]
	progPath := os.Args[2]
	etcdAddrs := os.Args[3:]
	cmdArgs := []string{envName, "0"}
	cmdArgs = append(cmdArgs, etcdAddrs...)

	mg, err := NewManger(progPath, cmdArgs)
	if err != nil {
		log.Fatal(err)
	}

	mg.Run()
}

func help() {
	info := `
Usage:%s [envName] [path] [etcdAddr...]
  envName  env namespace
  path     dir path
  etcdAddr etcd addr list
`
	fmt.Printf(info, os.Args[0])
}
