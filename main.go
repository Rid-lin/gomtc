package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

// NetFlow v5 implementation

type header struct {
	Version          uint16
	FlowRecords      uint16
	Uptime           uint32
	UnixSec          uint32
	UnixNsec         uint32
	FlowSeqNum       uint32
	EngineType       uint8
	EngineID         uint8
	SamplingInterval uint16
}

type binaryRecord struct {
	Ipv4SrcAddrInt uint32
	Ipv4DstAddrInt uint32
	Ipv4NextHopInt uint32
	InputSnmp      uint16
	OutputSnmp     uint16
	InPkts         uint32
	InBytes        uint32
	FirstInt       uint32
	LastInt        uint32
	L4SrcPort      uint16
	L4DstPort      uint16
	_              uint8
	TCPFlags       uint8
	Protocol       uint8
	SrcTos         uint8
	SrcAs          uint16
	DstAs          uint16
	SrcMask        uint8
	DstMask        uint8
	_              uint16
}

type decodedRecord struct {
	header
	binaryRecord

	Host              string
	SamplingAlgorithm uint8
	Ipv4SrcAddr       string
	Ipv4DstAddr       string
	Ipv4NextHop       string
	SrcHostName       string
	DstHostName       string
	Duration          uint16
}

func intToIPv4Addr(intAddr uint32) net.IP {
	return net.IPv4(
		byte(intAddr>>24),
		byte(intAddr>>16),
		byte(intAddr>>8),
		byte(intAddr))
}

func decodeRecord(header *header, binRecord *binaryRecord, remoteAddr *net.UDPAddr) decodedRecord {

	decodedRecord := decodedRecord{

		Host: remoteAddr.IP.String(),

		header: *header,

		binaryRecord: *binRecord,

		Ipv4SrcAddr: intToIPv4Addr(binRecord.Ipv4SrcAddrInt).String(),
		Ipv4DstAddr: intToIPv4Addr(binRecord.Ipv4DstAddrInt).String(),
		Ipv4NextHop: intToIPv4Addr(binRecord.Ipv4NextHopInt).String(),
		Duration:    uint16((binRecord.LastInt - binRecord.FirstInt) / 1000),
	}

	// LookupAddr
	decodedRecord.SrcHostName = lookUpWithCache(decodedRecord.Ipv4SrcAddr)
	decodedRecord.DstHostName = lookUpWithCache(decodedRecord.Ipv4DstAddr)

	// decode sampling info
	decodedRecord.SamplingAlgorithm = uint8(0x3 & (decodedRecord.SamplingInterval >> 14))
	decodedRecord.SamplingInterval = 0x3fff & decodedRecord.SamplingInterval

	return decodedRecord
}

func decodeRecordToSquid(header *header, binRecord *binaryRecord, remoteAddr string) string {
	srcmac := make([]byte, 8)
	dstmac := make([]byte, 8)
	binary.BigEndian.PutUint16(srcmac, binRecord.SrcAs)
	binary.BigEndian.PutUint16(dstmac, binRecord.DstAs)
	// srcmac = srcmac[2:8]
	// dstmac = dstmac[2:8]
	var protocol string

	switch fmt.Sprintf("%v", binRecord.Protocol) {
	case "6":
		protocol = "TCP_PACKET"
	case "17":
		protocol = "UDP_PACKET"
	case "1":
		protocol = "ICMP_PACKET"

	default:
		protocol = "OTHER_PACKET"
	}

	message := fmt.Sprintf("%v.000 %6v %v %v/- %v HEAD %v:%v %v FIRSTUP_PARENT/%v packet_netflow/%v/:%v ",
		binRecord.FirstInt,                               // time
		binRecord.FirstInt-binRecord.LastInt,             //delay
		intToIPv4Addr(binRecord.Ipv4DstAddrInt).String(), // dst ip
		protocol,          // protocol
		binRecord.InBytes, // size
		intToIPv4Addr(binRecord.Ipv4SrcAddrInt).String(), //src ip
		binRecord.L4SrcPort,               // src port
		net.HardwareAddr(srcmac).String(), // srcmac
		remoteAddr,                        // routerIP
		net.HardwareAddr(dstmac).String(), //dstmac
		binRecord.L4DstPort)               //dstport
	return message
}

func pipeOutputToStdout(outputChannel chan decodedRecord) {
	var record decodedRecord
	for {
		record = <-outputChannel
		out, _ := json.Marshal(record)
		fmt.Println(string(out))
	}
}

func pipeOutputToStdoutForSquid(outputChannel chan decodedRecord) {
	var record decodedRecord
	for {
		record = <-outputChannel
		message := decodeRecordToSquid(&record.header, &record.binaryRecord, record.Host)
		fmt.Printf("\n%s", message)
	}
}

type cacheRecord struct {
	Hostname string
	timeout  time.Time
}

var (
	cache      = map[string]cacheRecord{}
	cacheMutex = sync.RWMutex{}
)

func lookUpWithCache(ipAddr string) string {
	hostname := ipAddr
	cacheMutex.Lock()
	hostnameFromCache := cache[ipAddr]
	cacheMutex.Unlock()
	if (hostnameFromCache == cacheRecord{} || time.Now().After(hostnameFromCache.timeout)) {
		hostTemp, err := net.LookupAddr(ipAddr)
		if err == nil {
			if len(hostTemp) > 0 {
				hostname = hostTemp[0]
			}
		}
		cacheMutex.Lock()
		cache[ipAddr] = cacheRecord{hostname, time.Now().AddDate(0, 0, 1)}
		cacheMutex.Unlock()
	} else {
		hostname = hostnameFromCache.Hostname
	}
	return hostname
}

func formatLineProtocolForUDP(record decodedRecord) []byte {
	return []byte(fmt.Sprintf("netflow,host=%s,srcAddr=%s,dstAddr=%s,srcHostName=%s,dstHostName=%s,protocol=%d,srcPort=%d,dstPort=%d,input=%d,output=%d inBytes=%d,inPackets=%d,duration=%d %d",
		record.Host, record.Ipv4SrcAddr, record.Ipv4DstAddr, record.SrcHostName, record.DstHostName, record.Protocol, record.L4SrcPort, record.L4DstPort, record.InputSnmp, record.OutputSnmp,
		record.InBytes, record.InPkts, record.Duration,
		uint64((uint64(record.UnixSec)*uint64(1000000000))+uint64(record.UnixNsec))))
}

func pipeOutputToUDPSocket(outputChannel chan decodedRecord, targetAddr string) {
	/* Setting-up the socket to send data */

	remote, err := net.ResolveUDPAddr("udp", targetAddr)
	if err != nil {
		log.Printf("Name resolution failed: %v\n", err)
	} else {

		for {
			connection, err := net.DialUDP("udp", nil, remote)
			if err != nil {
				log.Printf("Connection failed: %v\n", err)
			} else {
				defer connection.Close()
				var record decodedRecord
				for {
					record = <-outputChannel
					var buf = formatLineProtocolForUDP(record)
					// message := string(buf)
					// log.Infof("%v", message)
					conn := connection
					err = conn.SetDeadline(time.Now().Add(3 * time.Second))
					if err != nil {
						log.Errorf("Error SetDeadline: %v", err)
						break
					}
					_, err := conn.Write(buf)
					if err != nil {
						log.Errorf("Send Error: %v", err)
						break
					}
				}
			}
		}
	}
}

func handlePacket(buf *bytes.Buffer, remoteAddr *net.UDPAddr, outputChannel chan decodedRecord) {
	header := header{}
	err := binary.Read(buf, binary.BigEndian, &header)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {

		for i := 0; i < int(header.FlowRecords); i++ {
			record := binaryRecord{}
			err := binary.Read(buf, binary.BigEndian, &record)
			if err != nil {
				log.Printf("binary.Read failed: %v\n", err)
				break
			}

			decodedRecord := decodeRecord(&header, &record, remoteAddr)
			outputChannel <- decodedRecord
		}
	}
}

func getExitSignalsChannel() chan os.Signal {

	c := make(chan os.Signal, 1)
	signal.Notify(c,
		// https://www.gnu.org/software/libc/manual/html_node/Termination-Signals.html
		syscall.SIGTERM, // "the normal way to politely ask a program to terminate"
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGQUIT, // Ctrl-\
		// syscall.SIGKILL, // "always fatal", "SIGKILL and SIGSTOP may not be caught by a program"
		syscall.SIGHUP, // "terminal is disconnected"
	)
	return c

}

func main() {
	/* Parse command-line arguments */
	var (
		inSource               string
		outMethod              string
		outDestination         string
		receiveBufferSizeBytes int
		conn                   *net.UDPConn
		err                    error
	)
	flag.StringVar(&inSource, "in", "0.0.0.0:2055", "Address and port to listen NetFlow packets")
	flag.StringVar(&outMethod, "method", "stdout", "Output method: stdout, udp")
	flag.StringVar(&outDestination, "out", "", "Address and port of influxdb to send decoded data")
	flag.IntVar(&receiveBufferSizeBytes, "buffer", 212992, "Size of RxQueue, i.e. value for SO_RCVBUF in bytes")
	flag.Parse()

	/*Creating a channel to intercept the program end signal*/
	exitChan := getExitSignalsChannel()

	go func() {
		<-exitChan
		conn.Close()
		log.Println("Shutting down")
		os.Exit(0)

	}()

	/* Create output pipe */
	outputChannel := make(chan decodedRecord, 100)
	switch outMethod {
	case "stdout":
		go pipeOutputToStdout(outputChannel)
	case "udp":
		go pipeOutputToUDPSocket(outputChannel, outDestination)
	case "squid":
		go pipeOutputToStdoutForSquid(outputChannel)
	default:
		log.Fatalf("Unknown schema: %v\n", outMethod)

	}

	/* Start listening on the specified port */
	log.Infof("Start listening on %v and sending to %v %v", inSource, outMethod, outDestination)
	addr, err := net.ResolveUDPAddr("udp", inSource)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	for {
		conn, err = net.ListenUDP("udp", addr)
		if err != nil {
			log.Errorln(err)
		} else {
			err = conn.SetReadBuffer(receiveBufferSizeBytes)
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

						go handlePacket(stream, remote, outputChannel)
					}
				}
			}
		}

	}
}
