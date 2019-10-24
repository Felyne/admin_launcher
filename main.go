package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/howeyc/fsnotify"
)

var (
	Version   = ""
	BuildTime = ""
)

var (
	envName   string
	progPath  string
	etcdAddrs []string
)

var manager = NewProcessManager()
var updateChan = make(chan bool, 1)

func main() {
	if len(os.Args) < 4 {
		if len(os.Args) == 2 && os.Args[1] == "-v" {
			fmt.Printf("version: %s\nbuildTime: %s\n",
				Version, BuildTime)
		} else {
			help()
		}
		os.Exit(1)
	}
	envName = os.Args[1]
	progPath = os.Args[2]
	etcdAddrs = os.Args[3:]

	go updateProcess()
	go checkFsChange()
	for {
		updateChan <- true
		time.Sleep(10 * time.Second)
	}
}

func checkFsChange() {
	w, err := fsnotify.NewWatcher()
	if nil != err {
		os.Exit(1)
	}
	w.Watch(progPath)
	for {
		select {
		case ev := <-w.Event:
			{
				manager.Stop(ev.Name)
				manager.Start(ev.Name, getArgs()...)
				log.Printf("restart %s\n", ev.Name)
			}
		case <-w.Error:
			updateChan <- true
		}
	}
}

func updateProcess() {
	for {
		<-updateChan
		startExistProgram(progPath)
		manager.StopNonExistProgram()
	}
}

//微服务程序的命令行参数
func getArgs() []string {
	args := []string{envName, "0"}
	args = append(args, etcdAddrs...)
	return args
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

func startExistProgram(dirPath string) {
	files, err := ioutil.ReadDir(progPath)
	if nil != err {
		return
	}
	for _, f := range files {
		name := f.Name()
		realPath := path.Join(progPath, f.Name())
		go func() {
			err := manager.Start(realPath, getArgs()...)
			if err != nil && err != ErrStarted {
				log.Println(err)
			} else if nil == err {
				log.Printf("start prog %s \n", name)
			}
		}()
	}
}
