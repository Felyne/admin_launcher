package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/atomic"
)

type Manager struct {
	progPath       string   //存放程序文件的目录路径
	cmdArgs        []string //运行程序的命令行参数
	watcher        *fsnotify.Watcher
	processManager *ProcessManager
	updateChan     chan struct{}
	running        atomic.Bool
}

func NewManger(progPath string, cmdArgs []string) (*Manager, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := w.Add(progPath); err != nil {
		return nil, err
	}
	mg := &Manager{
		progPath:       progPath,
		cmdArgs:        cmdArgs,
		watcher:        w,
		processManager: NewProcessManager(),
		updateChan:     make(chan struct{}, 1),
	}

	return mg, nil
}

func (mg *Manager) Run() {
	//返回旧值
	if mg.running.Swap(true) {
		log.Println("already Run")
		return
	}
	go mg.checkFsChange()
	go mg.updateProcess()
	go mg.waitExit()
	for {
		if mg.isRunning() == false {
			break
		}
		mg.updateChan <- struct{}{}
		time.Sleep(15 * time.Second)
	}
}

func (mg *Manager) checkFsChange() {
	for {
		if mg.isRunning() == false {
			break
		}
		select {
		case ev := <-mg.watcher.Events:
			{
				if ev.Op&fsnotify.Create == fsnotify.Create {
					mg.processManager.Start(ev.Name, mg.cmdArgs...)
					log.Printf("start program %s\n", ev.Name)
				}
				if ev.Op&fsnotify.Write == fsnotify.Write {
					mg.processManager.Stop(ev.Name)
					mg.processManager.Start(ev.Name, mg.cmdArgs...)
					log.Printf("restart program %s\n", ev.Name)
				}
				if ev.Op&fsnotify.Remove == fsnotify.Remove ||
					ev.Op&fsnotify.Rename == fsnotify.Rename {
					mg.processManager.Stop(ev.Name)
					log.Printf("stop program %s\n", ev.Name)
				}

			}
		case <-mg.watcher.Errors:
			mg.updateChan <- struct{}{}
		}
	}
}

func (mg *Manager) updateProcess() {
	for {
		if mg.isRunning() == false {
			break
		}
		<-mg.updateChan
		mg.processManager.StopNonExistProgram()
		files, err := ioutil.ReadDir(mg.progPath)
		if err != nil {
			continue
		}
		for _, f := range files {
			name := f.Name()
			realPath := path.Join(mg.progPath, f.Name())
			filepath.Join()
			go func() {
				err := mg.processManager.Start(realPath, mg.cmdArgs...)
				if err == nil {
					log.Printf("start program %s\n", name)
					return
				}
				if err != ErrStarted {
					log.Println(err)
				}
			}()
		}
	}
}

func (mg *Manager) waitExit() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-sigs
	log.Println("catch signal", sig)
	if sig == syscall.SIGINT || sig == syscall.SIGTERM || sig == syscall.SIGQUIT {
		mg.stopRunning()
		mg.processManager.StopAll()
		os.Exit(0)
	}
}

func (mg *Manager) isRunning() bool {
	return mg.running.Load()
}

func (mg *Manager) stopRunning() {
	mg.running.Store(false)
}
