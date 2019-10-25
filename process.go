package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

var ErrStarted = errors.New("the program has started")

type ProcessManager struct {
	filePidMap map[string]int //程序的绝对路径和对应的进程pid
	mu         sync.Mutex
}

func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		filePidMap: make(map[string]int),
	}
}

//程序路径和它的命令行参数
func (pm *ProcessManager) Start(filePath string, argv ...string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	_, ok := pm.filePidMap[absPath]
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
	if err != nil {
		log.Println(err)
		return err
	}
	pm.filePidMap[absPath] = p.Pid
	go pm.waitExit(absPath, p)

	return nil
}

func (pm *ProcessManager) waitExit(filePath string, p *os.Process) {
	stat, err := p.Wait()
	if err != nil {
		log.Println(err)
		return
	}
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pid := pm.filePidMap[filePath]
	if pid == p.Pid {
		delete(pm.filePidMap, filePath)
		log.Printf("program %s exist %s\n",
			filePath, stat.String())
	}
}

func (pm *ProcessManager) Stop(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pid, ok := pm.filePidMap[absPath]
	if false == ok {
		return nil
	}
	p, err := os.FindProcess(pid)
	if err == nil {
		p.Signal(syscall.SIGTERM)
		p.Signal(syscall.SIGINT)
		go func() {
			time.Sleep(100 * time.Millisecond)
			p.Kill()
		}()
	}

	delete(pm.filePidMap, absPath)

	return nil
}

//停止文件路径不存在的程序
func (pm *ProcessManager) StopNonExist() {
	absPathList := pm.filePathList()
	for _, absPath := range absPathList {
		if _, err := os.Stat(absPath); err != nil {
			pm.Stop(absPath)
			log.Printf("stop program %s\n", absPath)
		}
	}
}

func (pm *ProcessManager) StopAll() {
	absPathList := pm.filePathList()
	for _, absPath := range absPathList {
		pm.Stop(absPath)
	}
}

func (pm *ProcessManager) filePathList() []string {
	var absPathList []string
	pm.mu.Lock()
	for absPath, _ := range pm.filePidMap {
		absPathList = append(absPathList, absPath)
	}
	pm.mu.Unlock()
	return absPathList
}
