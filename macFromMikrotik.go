package main

import (
	"strconv"
	"sync"
	"time"

	"github.com/go-routeros/routeros"
	log "github.com/sirupsen/logrus"
)

func (data *transport) GetInfo(request *request) ResponseType {
	var response ResponseType

	timeInt, err := strconv.ParseInt(request.Time, 10, 64)
	if err != nil {
		log.Errorf("Error parsing timeStamp(%v) from request:%v", timeInt, err)
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
	} else if ok {
		// TODO убрать
		log.Tracef("IP:%v to MAC:%v, hostname:%v, comment:%v", ipStruct.ip, ipStruct.mac, ipStruct.hostName, ipStruct.comment)
		response.Mac = ipStruct.mac
		response.IP = ipStruct.ip
		response.Hostname = ipStruct.hostName
		response.Comment = ipStruct.comment
	} else if !ok {
		// TODO Сделать чтобы информация о мак-адресе загружалась из роутера
		log.Tracef("IP:'%v' not find in table lease of router:'%v'", ipStruct.ip, cfg.MTAddr)
		response.Mac = request.IP
		response.IP = request.IP
	} else {
		log.Tracef("IP:'%v' not find in table lease of router:'%v'", ipStruct.ip, cfg.MTAddr)
		response.Mac = request.IP
		response.IP = request.IP
	}

	return response
}

/*
Jun 22 21:39:13 192.168.65.1 dhcp,info dhcp_lan deassigned 192.168.65.149 from 04:D3:B5:FC:E8:09
Jun 22 21:40:16 192.168.65.1 dhcp,info dhcp_lan assigned 192.168.65.202 to E8:6F:38:88:92:29
*/

func NewTransport(cfg *Config) *transport {
	return &transport{
		ipToMac:     make(map[string]LineOfData),
		renewOneMac: make(chan string, 100),
		GMT:         cfg.GMT,
	}
}

func (data *transport) getDataFromMT(c *routeros.Client) {
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
