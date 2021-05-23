package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

// func CheckPIDFile(filename string) error {
// 	// View file info
// 	if stat, err := os.Stat(filename); err != nil {
// 		// If it is not there, start
// 		if os.IsNotExist(err) {
// 			return nil
// 		}

// 		// If the time is more than 15 minutes - delete this file and run the program
// 	} else if time.Since(stat.ModTime()) > 15*time.Minute {

// 		if err := os.Remove(filename); err != nil {
// 			log.Errorf("Error remove file(%v):%v", filename, err)
// 		}
// 		if err := writePID(filename); err != nil {
// 			return err
// 		}

// 		return nil
// 		// If it is there and the time for its change is less than 15 minutes, do not start.

// 	}
// 	return fmt.Errorf("already running")
// }

// func writePID(filename string) error {
// 	file, err := os.Create(filename)
// 	if err != nil {
// 		return fmt.Errorf("Error open file(%v):%v", filename, err)
// 	}
// 	defer file.Close()
// 	_, err2 := file.Write([]byte(fmt.Sprint(os.Getpid())))
// 	if err2 != nil {
// 		return fmt.Errorf("Error write file=(%v), data=(%v):%v", filename, os.Getpid(), err)
// 	}
// 	return nil
// }

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

func (t *Transport) Exit() {
	<-t.exitChan
	t.stopReadFromUDP <- 1
	// t.clientROS.Close()
	if err := t.fileDestination.Sync(); err != nil {
		log.Error(err)
	}
	t.fileDestination.Close()
	t.conn.Close()
	log.Println("Shutting down")
	time.Sleep(5 * time.Second)
	os.Exit(0)

}

func (t *Transport) ReOpenLogAfterLogroatate() {
	<-t.newLogChan
	var err error
	t.Lock()
	log.Println("Received a signal from logrotate, close the file.")
	if err := t.fileDestination.Sync(); err != nil {
		log.Error(err)
	}
	t.fileDestination.Close()
	if !cfg.NoFlow {
		t.fileDestination, err = os.OpenFile(cfg.NameFileToLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fileDestination.Close()
			log.Fatalf("Error, the '%v' file could not be created (there are not enough premissions or it is busy with another program): %v", cfg.NameFileToLog, err)
		}
	}
	t.Unlock()
	log.Println("Opening a new file.")
	time.Sleep(2 * time.Second)
}

func getNewLogSignalsChannel() chan os.Signal {

	c := make(chan os.Signal, 1)
	signal.Notify(c,
		// https://www.gnu.org/software/libc/manual/html_node/Termination-Signals.html
		syscall.SIGHUP, // "terminal is disconnected"
	)
	return c

}
