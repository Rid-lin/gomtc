package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-routeros/routeros"
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

// func openOutputDevice(filename string) *bufio.Writer {
// 	if filename == "" {
// 		writer = bufio.NewWriter(os.Stdout)
// 		log.Debug("Output in os.Stdout")
// 		return writer

// 	} else {
// 		var err error
// 		filetDestination, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 		if err != nil {
// 			log.Errorf("Error, the '%v' file could not be created (there are not enough premissions or it is busy with another program): %v", filename, err)
// 			writer = bufio.NewWriter(os.Stdout)
// 			filetDestination.Close()
// 			log.Debug("Output in os.Stdout with error open file")
// 			return writer

// 		} else {
// 			// defer filetDestination.Close()
// 			writer = bufio.NewWriter(filetDestination)
// 			log.Debugf("Output in file (%v)(%v)", filename, filetDestination)
// 			return writer

// 		}
// 	}

// }

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

func (data *transport) decodeRecordToSquid(record *decodedRecord, cfg *Config) string {
	binRecord := record.binaryRecord
	header := record.header
	remoteAddr := record.Host
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
		dstmac := data.GetInfo(&request{
			IP:   intToIPv4Addr(binRecord.Ipv4DstAddrInt).String(),
			Time: fmt.Sprint(header.UnixSec)}).Mac
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
		message = fmt.Sprintf("%v.000 %6v %v %v/- %v HEAD %v:%v %v FIRSTUP_PARENT/%v packet_netflow_inverse/%v/:%v ",
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

func (data *transport) pipeOutputToStdoutForSquid(outputChannel chan decodedRecord, filetDestination *os.File, cfg *Config) {
	var record decodedRecord
	for {
		record = <-outputChannel
		message := data.decodeRecordToSquid(&record, cfg)
		message = filtredMessage(message, cfg.IgnorList)
		if message == "" {
			continue
		}
		log.Tracef("Added to log:%v", message)
		if _, err := filetDestination.WriteString(message); err != nil {
			log.Errorf("Error writing data buffer:%v", err)
		}
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

type Cache struct {
	cache map[string]cacheRecord
	sync.RWMutex
}

var (
	cache Cache
	// cache      = map[string]cacheRecord{}
	// cacheMutex = sync.RWMutex{}
	// writer           *bufio.Writer
	filetDestination *os.File
)

func lookMacUpWithCache(timeInt uint32, ipAddr, addrMacFromSyslog string) string {
	var hostname string
	cache.Lock()
	hostnameFromCache := cache.cache[ipAddr]
	cache.Unlock()
	if (hostnameFromCache == cacheRecord{} || time.Now().After(hostnameFromCache.timeout)) {
		hostname = getMac(timeInt, ipAddr, addrMacFromSyslog)
		cache.Lock()
		cache.cache[ipAddr] = cacheRecord{hostname, time.Now().Add(1 * time.Minute)}
		cache.Unlock()
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
		// syscall.SIGHUP, // "terminal is disconnected"
	)
	return c

}

// func getNewLogSignalsChannel() chan os.Signal {

// 	c := make(chan os.Signal, 1)
// 	signal.Notify(c,
// 		// https://www.gnu.org/software/libc/manual/html_node/Termination-Signals.html
// 		syscall.SIGHUP, // "terminal is disconnected"
// 	)
// 	return c

// }

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

// func getMac(timeInt uint32, ip, addrMacFromSyslog string) string {
// 	time := fmt.Sprint(timeInt)
// 	URL := fmt.Sprintf("%v/getmac?ip=%v&time=%v", addrMacFromSyslog, ip, time)
// 	client := http.Client{}
// 	resp, err := client.Get(URL)
// 	if err != nil {
// 		log.Warning(err)
// 		return "00:00:00:00:00:00"
// 	}
// 	var response Response
// 	// var result map[string]interface{}
// 	err2 := json.NewDecoder(resp.Body).Decode(&response)
// 	if err2 != nil {
// 		log.Errorf("Error Decode JSON(%v):%v", resp.Body, err2)
// 		return "00:00:00:00:00:00"
// 	} else if response.Mac == "" {
// 		return "00:00:00:00:00:00"

// 	} else {
// 		return response.Mac

// 	}
// }

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
	addrMacFromSyslog   string     `yaml:"addrMacFromSyslog" toml:"addrmacfromsyslog" env:"ADDR_M4S"`
	outMethod           string     `yaml:"outMethod" toml:"outmethod" env:"OUT_METHOD"`
	outDestination      string     `yaml:"outDestination" toml:"outdestination" env:"OUT_DESTINATION"`
	BindAddr            string     `yaml:"BindAddr" toml:"bindaddr" env:"ADDR_M4M" envdefault:":3030"`
	MTAddr              string     `yaml:"MTAddr" toml:"mtaddr" env:"ADDR_MT"`
	MTUser              string     `yaml:"MTUser" toml:"mtuser" env:"USER_MT"`
	MTPass              string     `yaml:"MTPass" toml:"mtpass" env:"PASS_MT"`
	GMT                 string     `yaml:"GMT" toml:"gmt" env:"GMT"`
	// properties             string
	Interval               string
	receiveBufferSizeBytes int  `yaml:"receiveBufferSizeBytes" toml:"receiveBufferSizeBytes" env:"GONFLUX_BUFSIZE"`
	useTLS                 bool `yaml:"tls" toml:"tls" env:"TLS"`
	// async                  bool `yaml:"async" toml:"async" env:"ASYNC_MT"`
}

var (
	cfg Config
	// SubNets, IgnorList arrayFlags
)

func (data *transport) GetInfo(request *request) ResponseType {
	var response ResponseType

	timeInt, err := strconv.ParseInt(request.Time, 10, 64)
	if err != nil {
		log.Errorf("Error parsing timeStamp(%v) from request:%v", timeInt, err)
		// response.Mac = "00:00:00:00:00:00"
		// return response
		//При невернозаданном времени убираем 30 секунд из текущего времени, чтобы была возможность идентифицировать IP адрес
		timeInt = time.Now().Add(-30 * time.Second).Unix()
	}
	request.timeInt = timeInt
	data.RLock()
	ipStruct, ok := data.ipToMac[request.IP]
	data.RUnlock()
	if ok && timeInt < ipStruct.timeoutInt {
		log.Tracef("IP:%v to MAC:%v, hostname:%v, comment:%v", ipStruct.ip, ipStruct.mac, ipStruct.hostName, ipStruct.comment)
		response.Mac = ipStruct.mac
		response.IP = ipStruct.ip
		response.Hostname = ipStruct.hostName
		response.Comment = ipStruct.comment
	} else {
		log.Tracef("IP:%v not find, requests from Mikrotik:%v", ipStruct.ip, cfg.MTAddr)
		data.renewOneMac <- request.IP
		data.RLock()
		ipStruct, ok = data.ipToMac[request.IP]
		data.RUnlock()
		if ok {
			log.Tracef("IP:%v added with MAC:%v, hostname:%v, comment:%v", ipStruct.ip, ipStruct.mac, ipStruct.hostName, ipStruct.comment)
			response.Mac = ipStruct.mac
			response.IP = ipStruct.ip
			response.Hostname = ipStruct.hostName
			response.Comment = ipStruct.comment
		}
	}

	return response
}

/*
Jun 22 21:39:13 192.168.65.1 dhcp,info dhcp_lan deassigned 192.168.65.149 from 04:D3:B5:FC:E8:09
Jun 22 21:40:16 192.168.65.1 dhcp,info dhcp_lan assigned 192.168.65.202 to E8:6F:38:88:92:29
*/

func NewTransport(cfg *Config) *transport {
	return &transport{
		// mapTable: make(map[string][]lineOfLog),
		ipToMac:     make(map[string]LineOfData),
		renewOneMac: make(chan string, 100),
		GMT:         cfg.GMT,
	}
}

func (data *transport) getDataFromMT(c *routeros.Client) {
	var ip string
	go func() {
		ip = <-data.renewOneMac
		reply, err := c.Run("/ip/dhcp-server/lease/print", "?status=bound", "?disabled=false")
		if err != nil {
			log.Error(err)
		}
		fmt.Print(reply)
		var lineOfData LineOfData
		for _, re := range reply.Re {
			if re.Map["active-address"] != ip {
				continue
			}
			lineOfData.ip = re.Map["active-address"]
			lineOfData.mac = re.Map["active-mac-address"]
			lineOfData.timeout = re.Map["expires-after"]
			lineOfData.hostName = re.Map["host-name"]
			lineOfData.comment = re.Map["comment"]
			//Вычисляем время когда закончится аренда адреса
			timeStr, err := time.ParseDuration(lineOfData.timeout)
			if err != nil {
				timeStr = 10 * time.Second
			}
			// Записываем в переменную для дальнейшего быстрого сравнения
			lineOfData.timeoutInt = time.Now().Add(timeStr).Unix()

			data.RLock()
			data.ipToMac[lineOfData.ip] = lineOfData
			data.RUnlock()
		}
		if lineOfData.mac == "" {
			reply, err := c.Run("/ip/arp/print")
			if err != nil {
				log.Error(err)
			}
			fmt.Print(reply)
			var lineOfData LineOfData
			for _, re := range reply.Re {
				if re.Map["address"] != ip {
					continue
				}
				lineOfData.ip = re.Map["address"]
				lineOfData.mac = re.Map["mac-address"]
				lineOfData.timeoutInt = time.Now().Add(1 * time.Minute).Unix()

				data.RLock()
				data.ipToMac[lineOfData.ip] = lineOfData
				data.RUnlock()
			}

		}
	}()

	for {
		var lineOfData LineOfData
		reply, err := c.Run("/ip/arp/print")
		if err != nil {
			log.Error(err)
		}
		for _, re := range reply.Re {
			lineOfData.ip = re.Map["address"]
			lineOfData.mac = re.Map["mac-address"]
			lineOfData.timeoutInt = time.Now().Add(1 * time.Minute).Unix()

			data.Lock()
			data.ipToMac[lineOfData.ip] = lineOfData
			data.Unlock()

		}
		reply2, err2 := c.Run("/ip/dhcp-server/lease/print", "?status=bound", "?disabled=false")
		if err2 != nil {
			log.Error(err)
		}
		for _, re := range reply2.Re {
			lineOfData.ip = re.Map["active-address"]
			lineOfData.mac = re.Map["active-mac-address"]
			lineOfData.timeout = re.Map["expires-after"]
			lineOfData.hostName = re.Map["host-name"]
			lineOfData.comment = re.Map["comment"]
			//Вычисляем время когда закончится аренда адреса
			timeStr, err := time.ParseDuration(lineOfData.timeout)
			if err != nil {
				timeStr = 10 * time.Second
			}
			// Записываем в переменную для дальнейшего быстрого сравнения
			lineOfData.timeoutInt = time.Now().Add(timeStr).Unix()

			data.Lock()
			data.ipToMac[lineOfData.ip] = lineOfData
			data.Unlock()

		}
		var interval time.Duration
		interval, err = time.ParseDuration(cfg.Interval)
		if err != nil {
			interval = 10 * time.Minute
		}
		time.Sleep(interval)

	}
}

func handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w,
			`<html>
			<head>
			<title>go-macfrommikrotik</title>
			</head>
			<body>
			Более подробно на https://github.com/Rid-lin/go-macfrommikrotik
			</body>
			</html>
			`)
	}
}

func (data *transport) getmacHandler() http.HandlerFunc {
	var (
		request  request
		Response ResponseType
	)

	return func(w http.ResponseWriter, r *http.Request) {
		request.Time = r.URL.Query().Get("time")
		request.IP = r.URL.Query().Get("ip")
		Response = data.GetInfo(&request)
		log.Debugf(" | Request:'%v','%v' response:'%v'", request.Time, request.IP, Response.Mac)
		responseJSON, err := json.Marshal(Response)
		if err != nil {
			log.Errorf("Error Marshaling mac'%v'to JSON:'%v'", Response.Mac, err)
		}
		// fmt.Fprint(w, mac)
		_, err2 := w.Write(responseJSON)
		if err2 != nil {
			log.Errorf("Error send response:%v", err2)
		}
	}
}

type request struct {
	Time,
	IP string
	timeInt int64
}

type ResponseType struct {
	IP       string `JSON:"IP"`
	Mac      string `JSON:"Mac"`
	Hostname string `JSON:"Hostname"`
	Comment  string `JSON:"Comment"`
}

type LineOfData struct {
	ip,
	mac,
	timeout,
	hostName,
	comment string
	timeoutInt int64
}

type transport struct {
	ipToMac map[string]LineOfData
	// mapTable map[string][]lineOfLog
	GMT         string
	renewOneMac chan string
	sync.RWMutex
}

func dial(cfg *Config) (*routeros.Client, error) {
	if cfg.useTLS {
		return routeros.DialTLS(cfg.MTAddr, cfg.MTUser, cfg.MTPass, nil)
	}
	return routeros.Dial(cfg.MTAddr, cfg.MTUser, cfg.MTPass)
}

func newConfig(configFilename string) *Config {
	/* Parse command-line arguments */
	flag.StringVar(&cfg.addrMacFromSyslog, "port", "localhost:3030", "Address for service mac-address determining")
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
	err := cleanenv.ReadConfig(configFilename, &cfg)
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

	return &cfg
}

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
		log.Errorf("Error open Syslog file:%v", err)
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

	// newLogChan := getNewLogSignalsChannel()

	// go func() {
	// 	<-newLogChan
	// 	log.Println("Received a signal from logrotate, close the file.")
	// 	writer.Flush()
	// 	filetDestination.Close()
	// 	log.Println("Opening a new file.")
	// 	time.Sleep(2 * time.Second)
	// 	writer = openOutputDevice(cfg.NameFileToLog)

	// }()

	/* Create output pipe */
	outputChannel := make(chan decodedRecord, 100)
	switch cfg.outMethod {
	case "stdout":
		go pipeOutputToStdout(outputChannel)
	case "udp":
		go pipeOutputToUDPSocket(outputChannel, cfg.outDestination)
	case "squid":
		go data.pipeOutputToStdoutForSquid(outputChannel, filetDestination, cfg)
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

						go handlePacket(stream, remote, outputChannel, cfg)
					}
				}
			}
		}

	}
}
