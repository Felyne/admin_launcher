package main

import (
	"io/ioutil"
	"log"
	"path"

	"github.com/howeyc/fsnotify"
)

type Manager struct {
	progPath       string
	cmdArgs        []string
	watcher        *fsnotify.Watcher
	processManager *ProcessManager
}

func NewManger(progPath string, cmdArgs []string) (*Manager, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	manager := &Manager{
		progPath:       progPath,
		cmdArgs:        cmdArgs,
		watcher:        w,
		processManager: NewProcessManager(),
	}
	go manager.checkFsChange()
	return manager, nil
}

func (manager *Manager) Run() {

}

func (manager *Manager) checkFsChange() {
	manager.watcher.Watch(manager.progPath)
	for {
		select {
		case ev := <-manager.watcher.Event:
			{
				manager.processManager.Stop(ev.Name)
				manager.processManager.Start(ev.Name, manager.cmdArgs...)
				log.Printf("restart %s\n", ev.Name)
			}
		case <-manager.watcher.Error:
			updateChan <- true
		}
	}
}

func (manager *Manager) startExistProgram(dirPath string) {
	files, err := ioutil.ReadDir(progPath)
	if nil != err {
		return
	}
	for _, f := range files {
		name := f.Name()
		realPath := path.Join(progPath, f.Name())
		go func() {
			err := manager.processManager.Start(realPath, manager.cmdArgs...)
			if err != nil && err != ErrStarted {
				log.Println(err)
			} else if nil == err {
				log.Printf("start prog %s \n", name)
			}
		}()
	}
}
