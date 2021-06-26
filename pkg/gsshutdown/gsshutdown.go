package gsshutdown

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type GSS struct {
	ExitChan        chan os.Signal
	LogChan         chan os.Signal
	StopReadFromUDP chan uint8
}

func NewGSS(fnExitfunc func(ve interface{}), ve interface{}, fnRotate func(vr interface{}), vr interface{}) *GSS {
	var gss GSS
	ExitChan := getExitSignalsChannel()
	gss.ExitChan = ExitChan
	LogChan := getNewLogSignalsChannel()
	gss.LogChan = LogChan
	go gss.Exit(fnExitfunc, ve)
	go gss.GetSIGHUP(fnRotate, vr)
	return &gss
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

func (gss *GSS) Exit(fnExit func(ve interface{}), ve interface{}) {
	<-gss.ExitChan
	//gss.StopReadFromUDP <- 1
	fnExit(ve)
	log.Println("Shutting down")
	// time.Sleep(5 * time.Second)
	os.Exit(0)

}

func (gss *GSS) GetSIGHUP(fnRotate func(vr interface{}), vr interface{}) {
	<-gss.LogChan
	fnRotate(vr)
}

func getNewLogSignalsChannel() chan os.Signal {

	c := make(chan os.Signal, 1)
	signal.Notify(c,
		// https://www.gnu.org/software/libc/manual/html_node/Termination-Signals.html
		syscall.SIGHUP, // "terminal is disconnected"
	)
	return c

}
