package main

import (
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Manager struct {
	progPath       string   //存放程序文件的目录路径
	cmdArgs        []string //运行程序的命令行参数
	watcher        *fsnotify.Watcher
	processManager *ProcessManager
	updateChan     chan struct{}
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
	go mg.checkFsChange()
	go mg.updateProcess()
	for {
		mg.updateChan <- struct{}{}
		time.Sleep(15 * time.Second)
	}
}

func (mg *Manager) checkFsChange() {
	for {
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
