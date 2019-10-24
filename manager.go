package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

var ErrStarted = errors.New("the program has started")

type Manager struct {
	filePidMap map[string]int //程序的绝对路径和对应的进程pid
	mu         sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		filePidMap: make(map[string]int),
	}
}

func (manager *Manager) Start(filePath string, argv ...string) error {
	absPath, err := filepath.Abs(filePath)
	if nil != err {
		return err
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	_, ok := manager.filePidMap[absPath]
	if ok {
		return ErrStarted
	}
	programName := filepath.Base(absPath)
	args := []string{programName}
	args = append(args, argv...)
	log.Println(strings.Join(args, " "))
	p, err := os.StartProcess(absPath, args, &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	})
	if nil != err {
		log.Println(err)
		return err
	}
	manager.filePidMap[absPath] = p.Pid
	go manager.waitExit(absPath, p)
	return nil
}

func (manager *Manager) waitExit(filePath string, p *os.Process) {
	stat, err := p.Wait()
	if nil != err {
		log.Println(err)
		return
	}
	manager.mu.Lock()
	defer manager.mu.Unlock()
	pid, _ := manager.filePidMap[filePath]
	if pid == p.Pid {
		delete(manager.filePidMap, filePath)
		log.Printf("prog %s exist %s\n", filePath, stat.String())
	}
}

func (manager *Manager) Stop(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if nil != err {
		return err
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	pid, ok := manager.filePidMap[absPath]
	if false == ok {
		return nil
	}
	func() {
		p, err := os.FindProcess(pid)
		if err != nil {
			return
		}
		p.Signal(syscall.SIGTERM)
		//p.Signal(syscall.SIGINT)
	}()

	delete(manager.filePidMap, absPath)
	return nil
}

//停止文件路径不存在的程序
func (manager *Manager) StopNonExistProgram() {
	absPathList := make([]string, 0)
	func() {
		manager.mu.Lock()
		defer manager.mu.Unlock()

		for absPath, _ := range manager.filePidMap {
			absPathList = append(absPathList, absPath)
		}

	}()

	for _, absPath := range absPathList {
		_, err := os.Stat(absPath)
		if nil != err {
			manager.Stop(absPath)
			log.Printf("stop prog %s\n", absPath)
		}
	}
}
