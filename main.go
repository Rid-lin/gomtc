package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
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

func openOutputDevice(filename string) *bufio.Writer {
	if filename == "" {
		writer = bufio.NewWriter(os.Stdout)
		log.Debug("Output in os.Stdout")
		return writer

	} else {
		var err error
		filetDestination, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Errorf("Error, the '%v' file could not be created (there are not enough premissions or it is busy with another program): %v", filename, err)
			writer = bufio.NewWriter(os.Stdout)
			filetDestination.Close()
			log.Debug("Output in os.Stdout with error open file")
			return writer

		} else {
			// defer filetDestination.Close()
			writer = bufio.NewWriter(filetDestination)
			log.Debugf("Output in file (%v)(%v)", filename, filetDestination)
			return writer

		}
	}

}

func intToIPv4Addr(intAddr uint32) net.IP {
	return net.IPv4(
		byte(intAddr>>24),
		byte(intAddr>>16),
		byte(intAddr>>8),
		byte(intAddr))
}

func decodeRecord(header *header, binRecord *binaryRecord, remoteAddr *net.UDPAddr, cfg *Config) decodedRecord {

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
	// decodedRecord.SrcHostName = lookUpWithCache(decodedRecord.Ipv4SrcAddr)
	// decodedRecord.DstHostName = lookMacUpWithCache(header.UnixSec, decodedRecord.Ipv4DstAddr, cfg.addrMacFromSyslog)

	// decode sampling info
	decodedRecord.SamplingAlgorithm = uint8(0x3 & (decodedRecord.SamplingInterval >> 14))
	decodedRecord.SamplingInterval = 0x3fff & decodedRecord.SamplingInterval

	return decodedRecord
}

func decodeRecordToSquid(header *header, binRecord *binaryRecord, remoteAddr string, cfg *Config) string {
	srcmacB := make([]byte, 8)
	dstmacB := make([]byte, 8)
	binary.BigEndian.PutUint16(srcmacB, binRecord.SrcAs)
	binary.BigEndian.PutUint16(dstmacB, binRecord.DstAs)
	// srcmac = srcmac[2:8]
	// dstmac = dstmac[2:8]

	var protocol, message string

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

	ok := cfg.CheckEntryInSubNet(intToIPv4Addr(binRecord.Ipv4DstAddrInt))
	ok2 := cfg.CheckEntryInSubNet(intToIPv4Addr(binRecord.Ipv4SrcAddrInt))

	if ok && !ok2 {
		dstmac := lookMacUpWithCache(header.UnixSec, intToIPv4Addr(binRecord.Ipv4DstAddrInt).String(), cfg.addrMacFromSyslog)
		message = fmt.Sprintf("%v.000 %6v %v %v/- %v HEAD %v:%v %v FIRSTUP_PARENT/%v packet_netflow/%v/:%v ",
			header.UnixSec,                                   // time
			binRecord.LastInt-binRecord.FirstInt,             //delay
			intToIPv4Addr(binRecord.Ipv4DstAddrInt).String(), // dst ip
			protocol,          // protocol
			binRecord.InBytes, // size
			intToIPv4Addr(binRecord.Ipv4SrcAddrInt).String(), //src ip
			binRecord.L4SrcPort,                // src port
			dstmac,                             // dstmac
			remoteAddr,                         // routerIP
			net.HardwareAddr(srcmacB).String(), // srcmac
			binRecord.L4DstPort)                // dstport

	} else if !ok && ok2 {
		dstmac := lookMacUpWithCache(header.UnixSec, intToIPv4Addr(binRecord.Ipv4SrcAddrInt).String(), cfg.addrMacFromSyslog)
		message = fmt.Sprintf("%v.000 %6v %v %v/- %v HEAD %v:%v %v FIRSTUP_PARENT/%v packet_netflow_inverse//%v/:%v ",
			header.UnixSec,                                   // time
			binRecord.LastInt-binRecord.FirstInt,             //delay
			intToIPv4Addr(binRecord.Ipv4SrcAddrInt).String(), //src ip - Local
			protocol,          // protocol
			binRecord.InBytes, // size
			intToIPv4Addr(binRecord.Ipv4DstAddrInt).String(), // dst ip - Inet
			binRecord.L4SrcPort,                // src port
			dstmac,                             // dstmac
			remoteAddr,                         // routerIP
			net.HardwareAddr(srcmacB).String(), // srcmac
			binRecord.L4DstPort)                // dstport

	}
	return message
}

func (cfg *Config) CheckEntryInSubNet(ipv4addr net.IP) bool {
	for _, subNet := range cfg.SubNets {
		ok, err := checkIP(subNet, ipv4addr)
		if err != nil { // если ошибка, то следующая строка
			log.Error("Error while determining the IP subnet address:", err)
			return false

		}
		if ok {
			return true
		}
	}

	return false
}

func checkIP(subnet string, ipv4addr net.IP) (bool, error) {
	_, netA, err := net.ParseCIDR(subnet)
	if err != nil {
		return false, err
	}

	return netA.Contains(ipv4addr), nil
}

func pipeOutputToStdout(outputChannel chan decodedRecord) {
	var record decodedRecord
	for {
		record = <-outputChannel
		out, _ := json.Marshal(record)
		fmt.Println(string(out))
	}
}

func pipeOutputToStdoutForSquid(outputChannel chan decodedRecord, writer *bufio.Writer, cfg *Config) {
	var record decodedRecord
	for {
		record = <-outputChannel
		message := decodeRecordToSquid(&record.header, &record.binaryRecord, record.Host, cfg)
		message = filtredMessage(message, cfg.IgnorList)
		if message == "" {
			continue
		}
		log.Tracef("Added to log:%v", message)
		fmt.Fprintf(writer, "\n%s", message)
	}
}

func filtredMessage(message string, IgnorList arrayFlags) string {
	for _, ignorStr := range cfg.IgnorList {
		if strings.Contains(message, ignorStr) {
			log.Tracef("Line of log :%v contains ignorstr:%v, skipping...", message, ignorStr)
			return ""
		}
	}
	return message
}

type cacheRecord struct {
	Hostname string
	timeout  time.Time
}

var (
	cache            = map[string]cacheRecord{}
	cacheMutex       = sync.RWMutex{}
	writer           *bufio.Writer
	filetDestination *os.File
)

func lookMacUpWithCache(timeInt uint32, ipAddr, addrMacFromSyslog string) string {
	var hostname string
	cacheMutex.Lock()
	hostnameFromCache := cache[ipAddr]
	cacheMutex.Unlock()
	if (hostnameFromCache == cacheRecord{} || time.Now().After(hostnameFromCache.timeout)) {
		hostname = getMac(timeInt, ipAddr, addrMacFromSyslog)
		cacheMutex.Lock()
		cache[ipAddr] = cacheRecord{hostname, time.Now().Add(10 * time.Minute)}
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

func handlePacket(buf *bytes.Buffer, remoteAddr *net.UDPAddr, outputChannel chan decodedRecord, cfg *Config) {
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

			decodedRecord := decodeRecord(&header, &record, remoteAddr, cfg)
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

type Response struct {
	Mac string `JSON:"Mac"`
}

func getMac(timeInt uint32, ip, addrMacFromSyslog string) string {
	time := fmt.Sprint(timeInt)
	URL := fmt.Sprintf("%v/getmac?ip=%v&time=%v", addrMacFromSyslog, ip, time)
	client := http.Client{}
	resp, err := client.Get(URL)
	if err != nil {
		log.Warning(err)
		return "00:00:00:00:00:00"
	}
	var response Response
	// var result map[string]interface{}
	err2 := json.NewDecoder(resp.Body).Decode(&response)
	if err2 != nil {
		log.Errorf("Error Decode JSON(%v):%v", resp.Body, err2)
		return "00:00:00:00:00:00"
	} else if response.Mac == "" {
		return "00:00:00:00:00:00"

	} else {
		return response.Mac

	}
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "List of strings"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type Config struct {
	SubNets             arrayFlags `yaml:"SubNets" toml:"subnets" env:"SUBNETS"`
	IgnorList           arrayFlags `yaml:"IgnorList" toml:"ignorlist" env:"IGNORLIST"`
	LogLevel            string     `yaml:"LogLevel" toml:"loglevel" env:"LOG_LEVEL"`
	ProcessingDirection string     `yaml:"ProcessingDirection" toml:"direct" env:"DIRECT" env-default:"both"`
	FlowAddr            string     `yaml:"FlowAddr" toml:"flowaddr" env:"FLOW_ADDR" env-default:"0.0.0.0:2055"`
	NameFileToLog       string     `yaml:"FileToLog" toml:"log" env:"FLOW_LOG"`
	// inSource               string     `yaml:"inSource" toml:"insource" env:"INSOURCE"`
	addrMacFromSyslog string `yaml:"addrMacFromSyslog" toml:"addrmacfromsyslog" env:"ADDR_M4S"`
	outMethod         string `yaml:"outMethod" toml:"outmethod" env:"OUT_METHOD"`
	outDestination    string `yaml:"outDestination" toml:"outdestination" env:"OUT_DESTINATION"`
	// configFilename         string
	receiveBufferSizeBytes int `yaml:"receiveBufferSizeBytes" toml:"receiveBufferSizeBytes" env:"GONFLUX_BUFSIZE"`
}

var (
	cfg Config
	// SubNets, IgnorList arrayFlags
)

func main() {
	var (
		conn           *net.UDPConn
		err            error
		configFilename string = "config.toml"
	)
	/* Parse command-line arguments */
	flag.StringVar(&cfg.addrMacFromSyslog, "port", "http://localhost:3030", "Address for service mac-address determining")
	flag.StringVar(&cfg.outMethod, "method", "stdout", "Output method: stdout, udp or squid")
	flag.StringVar(&cfg.outDestination, "out", "", "Address and port of influxdb to send decoded data")
	flag.IntVar(&cfg.receiveBufferSizeBytes, "buffer", 212992, "Size of RxQueue, i.e. value for SO_RCVBUF in bytes")
	flag.StringVar(&cfg.FlowAddr, "addr", "0.0.0.0:2055", "Address and port to listen NetFlow packets")
	flag.StringVar(&cfg.LogLevel, "loglevel", "info", "Log level")
	flag.Var(&cfg.SubNets, "subnet", "List of internal subnets")
	flag.Var(&cfg.IgnorList, "ignorlist", "List of ignored words/parameters per string")
	flag.StringVar(&cfg.ProcessingDirection, "direct", "both", "")
	flag.StringVar(&cfg.NameFileToLog, "log", "", "The file where logs will be written in the format of squid logs")

	flag.Parse()

	var config_source string
	err = cleanenv.ReadConfig(configFilename, &cfg)
	if err != nil {
		log.Warningf("No config file(%v) found: %v", configFilename, err)
		config_source = "ENV/CFG"
	} else {
		config_source = "CLI"
	}

	lvl, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Errorf("Error in determining the level of logs (%v). Installed by default = Info", cfg.LogLevel)
		lvl, _ = log.ParseLevel("info")
	}
	log.SetLevel(lvl)

	log.Debugf("Config read from %s: addrMacFromSyslog=(%s), configFilename=(%s), outDestination=(%s), receiveBufferSizeBytes=(%d), FlowAddr=(%s), LogLevel=(%s), SubNets=(%p), IgnorList=(%p), ProcessingDirection=(%s), NameFileToLog=(%s), ",
		config_source,
		cfg.addrMacFromSyslog,
		cfg.outMethod,
		cfg.outDestination,
		cfg.receiveBufferSizeBytes,
		cfg.FlowAddr,
		cfg.LogLevel,
		cfg.SubNets,
		cfg.IgnorList,
		cfg.ProcessingDirection,
		cfg.NameFileToLog)

	writer := openOutputDevice(cfg.NameFileToLog)

	/*Creating a channel to intercept the program end signal*/
	exitChan := getExitSignalsChannel()

	go func() {
		<-exitChan
		writer.Flush()
		conn.Close()
		log.Println("Shutting down")
		os.Exit(0)

	}()

	/* Create output pipe */
	outputChannel := make(chan decodedRecord, 100)
	switch cfg.outMethod {
	case "stdout":
		go pipeOutputToStdout(outputChannel)
	case "udp":
		go pipeOutputToUDPSocket(outputChannel, cfg.outDestination)
	case "squid":
		go pipeOutputToStdoutForSquid(outputChannel, writer, &cfg)
	default:
		log.Fatalf("Unknown schema: %v\n", cfg.outMethod)

	}

	/* Start listening on the specified port */
	log.Infof("Start listening on %v and sending to %v %v", cfg.FlowAddr, cfg.outMethod, cfg.outDestination)
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

						go handlePacket(stream, remote, outputChannel, &cfg)
					}
				}
			}
		}

	}
}
