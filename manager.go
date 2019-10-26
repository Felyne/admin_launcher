package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
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
	stopChan       chan struct{}
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
		stopChan:       make(chan struct{}),
	}

	return mg, nil
}

func (mg *Manager) Run() {
	//返回旧值
	if mg.running.Swap(true) {
		return
	}
	mg.review()
	mg.waitExit()
	mg.run()
}

func (mg *Manager) run() {
	for {
		select {
		case ev := <-mg.watcher.Events:
			{
				if ev.Op&fsnotify.Create == fsnotify.Create {
					mg.startProcess(ev.Name)
					log.Printf("start program %s\n", ev.Name)
				}
				if ev.Op&fsnotify.Write == fsnotify.Write {
					mg.stopProcess(ev.Name)
					mg.startProcess(ev.Name)
					log.Printf("restart program %s\n", ev.Name)
				}
				if ev.Op&fsnotify.Remove == fsnotify.Remove ||
					ev.Op&fsnotify.Rename == fsnotify.Rename {
					mg.stopProcess(ev.Name)
					log.Printf("stop program %s\n", ev.Name)
				}
			}
		case <-mg.watcher.Errors:
			mg.review()
		case <-time.After(15 * time.Second):
			mg.review()
		case <-mg.stopChan:
			mg.stopAllProcess()
			return
		}
	}
}

//复查
//停掉文件路径不存在的程序，启动文件存在但尚未启动的程序
func (mg *Manager) review() {
	mg.processManager.StopNonExist()
	files, err := ioutil.ReadDir(mg.progPath)
	if err != nil {
		return
	}
	for _, f := range files {
		name := f.Name()
		realPath := path.Join(mg.progPath, f.Name())
		err := mg.startProcess(realPath)
		if err == nil {
			log.Printf("start program %s\n", name)
		} else if err != ErrStarted {
			log.Println(err)
		}
	}
}

func (mg *Manager) startProcess(filePath string) error {
	return mg.processManager.Start(filePath, mg.cmdArgs...)
}

func (mg *Manager) stopProcess(filePath string) error {
	return mg.processManager.Stop(filePath)
}

func (mg *Manager) stopAllProcess() {
	mg.processManager.StopAll()
}

func (mg *Manager) waitExit() {
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT,
			syscall.SIGTERM, syscall.SIGQUIT)
		sig := <-ch
		log.Println("catch signal", sig)
		mg.stopChan <- struct{}{}
		mg.stopRunning()
	}()

}

func (mg *Manager) isRunning() bool {
	return mg.running.Load()
}

func (mg *Manager) stopRunning() {
	mg.running.Store(false)
}
