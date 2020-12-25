package main

import (
	"bytes"
	"net"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	var (
		conn           *net.UDPConn
		err            error
		configFilename string = "config.toml"
	)

	cfg := newConfig(configFilename)

	filetDestination, err = os.OpenFile(cfg.NameFileToLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		filetDestination.Close()
		log.Fatalf("Error, the '%v' file could not be created (there are not enough premissions or it is busy with another program): %v", cfg.NameFileToLog, err)
	}

	cache.cache = make(map[string]cacheRecord)

	/*Creating a channel to intercept the program end signal*/
	exitChan := getExitSignalsChannel()

	c, err := dial(cfg)
	if err != nil {
		log.Errorf("Error connect to %v:%v", cfg.MTAddr, err)
	}
	defer c.Close()

	data := NewTransport(cfg)
	go data.getDataFromMT(c)

	http.HandleFunc("/", handleIndex())
	http.HandleFunc("/getmac", data.getmacHandler())

	log.Infof("MacFromMikrotik-server listen %v", cfg.BindAddr)

	go func() {
		err := http.ListenAndServe(cfg.BindAddr, nil)
		if err != nil {
			log.Error("http-server returned error:", err)
		}
	}()

	go func() {
		<-exitChan
		c.Close()
		filetDestination.Close()
		conn.Close()
		log.Println("Shutting down")
		os.Exit(0)

	}()

	/* Create output pipe */
	outputChannel := make(chan decodedRecord, 100)

	go data.pipeOutputToStdoutForSquid(outputChannel, filetDestination, cfg)

	/* Start listening on the specified port */
	log.Infof("Start listening on %v", cfg.FlowAddr)
	addr, err := net.ResolveUDPAddr("udp", cfg.FlowAddr)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	for {
		conn, err = net.ListenUDP("udp", addr)
		if err != nil {
			log.Errorln(err)
		} else {
			err = conn.SetReadBuffer(cfg.receiveBufferSizeBytes)
			if err != nil {
				log.Errorln(err)
			} else {
				/* Infinite-loop for reading packets */
				for {
					buf := make([]byte, 4096)
					rlen, remote, err := conn.ReadFromUDP(buf)

					if err != nil {
						log.Errorf("Error: %v\n", err)
					} else {

						stream := bytes.NewBuffer(buf[:rlen])

						go handlePacket(stream, remote, outputChannel, cfg)
					}
				}
			}
		}

	}
}
