package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func writePID(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Error open PID-file(name=%v):%v", filename, err)
	}
	defer file.Close()
	_, err2 := file.Write([]byte(fmt.Sprint(os.Getpid())))
	if err2 != nil {
		return fmt.Errorf("Error write file=(%v), data=(%v):%v", filename, os.Getpid(), err)
	}
	return nil
}

func getExitSignalsChannel() chan os.Signal {

	c := make(chan os.Signal, 1)
	signal.Notify(c,
		// https://www.gnu.org/software/libc/manual/html_node/Termination-Signals.html
		syscall.SIGTERM, // "the normal way to politely ask a program to terminate"
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGQUIT, // Ctrl-\
		// syscall.SIGKILL, // "always fatal", "SIGKILL and SIGSTOP may not be caught by a program"
		// syscall.SIGHUP, // "terminal is disconnected"
	)
	return c

}

func (t *Transport) Exit(cfg *Config) {
	<-t.exitChan
	t.stopReadFromUDP <- 1
	if !cfg.NoFlow {
		if err := t.fileDestination.Sync(); err != nil {
			log.Errorf("File(%v) sync error:%v", t.fileDestination.Name(), err)
		}
		if err := t.fileDestination.Close(); err != nil {
			log.Errorf("File(%v) close error:%v", t.fileDestination.Name(), err)
		}
	}
	if err := os.Remove(t.pidfile); err != nil {
		log.Errorf("File (%v) deletion error:%v", t.pidfile, err)
	}
	log.Println("Shutting down")
	// time.Sleep(5 * time.Second)
	os.Exit(0)

}

func (t *Transport) ReOpenLogAfterLogroatate(cfg *Config) {
	<-t.newLogChan
	var err error
	t.Lock()
	log.Println("Received a signal from logrotate, close the file.")
	if err := t.fileDestination.Sync(); err != nil {
		log.Errorf("File(%v) sync error:%v", t.fileDestination.Name(), err)
	}
	if err := t.fileDestination.Close(); err != nil {
		log.Errorf("File(%v) close error:%v", t.fileDestination.Name(), err)
	}
	if !cfg.NoFlow {
		t.fileDestination, err = os.OpenFile(cfg.NameFileToLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fileDestination.Close()
			log.Fatalf("Error, the '%v' file could not be created (there are not enough premissions or it is busy with another program): %v", cfg.NameFileToLog, err)
		}
	}
	t.Unlock()
}

func getNewLogSignalsChannel() chan os.Signal {

	c := make(chan os.Signal, 1)
	signal.Notify(c,
		// https://www.gnu.org/software/libc/manual/html_node/Termination-Signals.html
		syscall.SIGHUP, // "terminal is disconnected"
	)
	return c

}
