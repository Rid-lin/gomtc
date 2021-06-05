package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var count Count

func (t *Transport) runOnce(cfg *Config) {
	p := parseType{}
	t.RLock()
	p.SSHCredentials = t.sshCredentials
	p.BlockAddressList = t.BlockAddressList
	p.QuotaType = t.QuotaType
	p.Location = t.Location
	t.RUnlock()

	t.readLog(cfg)

	setLastdates(1, 1, cfg)

	t.updateDevices()

	t.parseAllFilesAndCountingTraffic(cfg)
	t.totalTrafficСounting()

	t.writeToChasheData()
	t.updateAliases(p)
	t.checkQuotas()
	t.updateDevices()
	t.updateAliases(p)

	t.clearingCountedTraffic(cfg, cfg.LastDate)

	t.writeLog(cfg)

	count = Count{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	t.parseAllFilesAndCountingTrafficNew(cfg)
	// t.Print()
	t.setTimerParse(cfg.ParseDelay)
}

func (t *Transport) setTimerParse(IntervalStr string) {
	interval, err := time.ParseDuration(IntervalStr)
	if err != nil {
		t.timerParse = time.NewTimer(15 * time.Minute)
	} else {
		t.timerParse = time.NewTimer(interval)
	}
}

func (t *Transport) parseAllFilesAndCountingTraffic(cfg *Config) {
	// Получение текущего времени для расчёта времени работы
	cfg.startTime = time.Now()
	fmt.Printf("Parsing has started.\n")
	err := t.parseDirToMapOld(cfg)
	if err != nil {
		log.Error(err)
	}
	ExTime := time.Since(cfg.startTime)
	ExTimeInSec := uint64(ExTime.Seconds())
	if ExTimeInSec == 0 {
		ExTimeInSec = 1
	}
	cfg.endTime = time.Now() // Saves the current time to be inserted into the log table
	t.lastUpdated = time.Now()
	log.Infof("The parsing started at %v, ended at %v, and lasted %.3v seconds at a rate of %v lines per second.",
		cfg.startTime.In(cfg.Location).Format(cfg.dateTimeLayout),
		cfg.endTime.In(cfg.Location).Format(cfg.dateTimeLayout),
		ExTime.Seconds(),
		cfg.totalLineParsed/ExTimeInSec)
	fmt.Printf("The parsing started at %v, ended at %v, and lasted %.3v seconds at a rate of %v lines per second.\n",
		cfg.startTime.In(cfg.Location).Format(cfg.dateTimeLayout),
		cfg.endTime.In(cfg.Location).Format(cfg.dateTimeLayout),
		ExTime.Seconds(),
		cfg.totalLineParsed/ExTimeInSec)
}

func (t *Transport) clearingCountedTraffic(cfg *Config, timestamp int64) {
	cfg.LastDay = findOutTheCurrentDay(cfg.LastDate, cfg.Location)
	t.dataOld = MapOfReports{}
}

func (t *Transport) ClearDataOfLastDay(cfg *Config) {
	dateStr := time.Unix(cfg.LastDate, 0).In(cfg.Location).Format(cfg.dateLayout)
	cfg.LastDay = findOutTheCurrentDay(cfg.LastDate, cfg.Location)

	t.RLock()
	data := t.dataOld
	t.RUnlock()
	for key := range data {
		if key.DateStr == dateStr {
			delete(data, key)
			log.Tracef("Item key(%v) data(%v)(%v)(%v)) was deleted", key, data[key].VolumePerDay, data[key].Hits, data[key].VolumePerHour)
		}
	}
	t.Lock()
	t.dataOld = data
	t.Unlock()
	cfg.LastDate = cfg.LastDay
}

func (t *Transport) parseDirToMapOld(cfg *Config) error {
	// iteration over all files in a folder
	files, err := ioutil.ReadDir(cfg.LogPath)
	if err != nil {
		return err
	}
	SortFileByModTime(files)
	for _, file := range files {
		if err := t.parseFileToMapOld(file, cfg); err != nil {
			log.Error(err)
			continue
		}
		fmt.Printf("From file %v lines Read:%v/Parsed:%v/Added:%v/Skiped:%v/Error:%v\n",
			file.Name(),
			cfg.LineRead,
			cfg.LineParsed,
			cfg.LineAdded,
			cfg.LineSkiped,
			cfg.LineError)
		cfg.SumAndReset()
	}
	fmt.Printf("From all files lines Read:%v/Parsed:%v/Added:%v/Skiped:%v/Error:%v\n",
		// Lines read:%v, parsed:%v, lines added:%v lines skiped:%v lines error:%v",
		cfg.totalLineRead,
		cfg.totalLineParsed,
		cfg.totalLineAdded,
		cfg.totalLineSkiped,
		cfg.totalLineError)
	return nil
}

func (t *Transport) parseFileToMapOld(info os.FileInfo, cfg *Config) error {
	FullFileName := filepath.Join(cfg.LogPath, info.Name())
	file, err := os.Open(FullFileName)
	if err != nil {
		file.Close()
		return fmt.Errorf("Error opening squid log file(FullFileName):%v", err)
	}
	defer file.Close()
	// Проверить является ли файл архивом
	_, errGZip := gzip.NewReader(file)
	// Если файл читается с ошибками и это не ошибка gzip.ErrHeader, то возвращаем ошибку
	if errGZip != nil && errGZip != gzip.ErrHeader {
		return errGZip
		// Если нет ошибок, то извлекаем файл во временную папку и передаём в след.шаг имя файла
	} else if errGZip == nil {
		dir, err := ioutil.TempDir("", "gomtc")
		if err != nil {
			log.Errorf("Error extracting file to temporary folder:%v", err)
		}
		defer os.RemoveAll(dir) // очистка
		FileName, err := unGzip(FullFileName, dir)
		if err != nil {
			return err
		}
		FullFileName = FileName
		// FullFileName = filepath.Join(path, FileName)
	}
	// Если ошибка gzip.ErrHeader, то обрабатываем как текстовый файл.
	file.Close()
	file, err = os.Open(FullFileName)
	if err != nil {
		file.Close()
		return fmt.Errorf("Error opening squid log file(FullFileName):%v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if err := t.parseLogToArrayByLineOld(scanner, cfg); err != nil {
		return err
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}
	return nil
}

func (t *Transport) parseLogToArrayByLineOld(scanner *bufio.Scanner, cfg *Config) error {
	for scanner.Scan() { // We go through the entire file to the end
		cfg.LineRead++
		line := scanner.Text() // get the text from the line, for simplicity
		if line == "" {
			cfg.LineSkiped++
			continue
		}
		line = replaceQuotes(line)
		lineOut, err := parseLineToStruct(line, cfg)
		if err != nil {
			cfg.LineError++
			log.Warningf("%v", err)
			continue
		}
		cfg.LineParsed++
		if cfg.LastDate > lineOut.timestamp {
			log.Tracef("line(%v) too old\r", line)
			cfg.LineSkiped++
			continue
		} else if cfg.LastDate < lineOut.timestamp {
			cfg.LastDate = lineOut.timestamp
		}

		// The main function of filling the database
		// Adding a row to the database for counting traffic
		// t.addLineOutToMapOfReportsSuperNew(&lineOut)
		// t.addLineOutToMapOfReports(&lineOut)
		t.addLineOutToMapOfReportsOld(&lineOut, cfg)

		cfg.LineAdded++
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Errorf("Error scanner :%v", err)
		return err
	}
	return nil
}

func replaceQuotes(lineOld string) string {
	lineNew := strings.ReplaceAll(lineOld, "'", "&quot")
	line := strings.ReplaceAll(lineNew, `"`, "&quot")
	return line
}

func squidDateToINT64(squidDate string) (timestamp, nsec int64, err error) {
	timestampStr := strings.Split(squidDate, ".")
	timestampStrSec := timestampStr[0]
	if len(timestampStrSec) > 10 {
		timestampStrSec = timestampStrSec[len(timestampStrSec)-10:]
	}
	timestamp, err = strconv.ParseInt(timestampStrSec, 10, 64)
	if err != nil {
		return 0, 0, err
	}
	if len(timestampStr) > 1 {
		nsec, err = strconv.ParseInt(timestampStr[1], 10, 64)
		if err != nil {
			return 0, 0, err
		}
	}
	return
}

func parseLineToStruct(line string, cfg *Config) (lineOfLogType, error) {
	var l lineOfLogType
	var err error
	valueArray := strings.Fields(line) // split into fields separated by a space to parse into a structure
	if len(valueArray) < 10 {          // check the length of the resulting array to make sure that the string is parsed normally and there are no errors in its format
		return l, fmt.Errorf("Error, string(%v) is not line of Squid-log", line) // If there is an error, then we stop working to avoid unnecessary transformations
	}
	l.date = valueArray[0]
	l.timestamp, l.nsec, err = squidDateToINT64(l.date)
	if err != nil {
		return lineOfLogType{}, err
	}
	timeUnix := time.Unix(l.timestamp, 0)
	l.year = timeUnix.Year()
	l.month = timeUnix.Month()
	l.day = timeUnix.Day()
	l.hour = timeUnix.Hour()
	l.minute = timeUnix.Minute()
	l.ipaddress = valueArray[2]
	l.httpstatus = valueArray[3]
	sizeInBytes, err := strconv.ParseUint(valueArray[4], 10, 64)
	if err != nil {
		sizeInBytes = 0
	}
	l.sizeInBytes = sizeInBytes
	l.method = valueArray[5]
	l.siteName = valueArray[6]
	l.login = valueArray[7]
	l.mime = valueArray[9]
	if len(valueArray) > 10 {
		l.hostname = valueArray[10]
	} else {
		l.hostname = ""
	}
	if len(valueArray) > 11 {
		l.comments = strings.Join(valueArray[11:], " ")
	} else {
		l.comments = ""
	}
	return l, nil
}

func determiningAlias(value lineOfLogType) string {
	alias := value.alias
	if alias == "" {
		if value.login == "" || value.login == "-" {
			alias = value.ipaddress
		} else {
			alias = value.login
		}
	}
	return alias
}

func SortFileByModTime(files []os.FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Unix() < files[j].ModTime().Unix()
	})
}

func (count *Count) SumAndReset() {
	count.totalLineRead = count.totalLineRead + count.LineRead
	count.totalLineParsed = count.totalLineParsed + count.LineParsed
	count.totalLineAdded = count.totalLineAdded + count.LineAdded
	count.totalLineSkiped = count.totalLineSkiped + count.LineSkiped
	count.totalLineError = count.totalLineError + count.LineError
	count.LineParsed = 0
	count.LineSkiped = 0
	count.LineAdded = 0
	count.LineRead = 0
	count.LineError = 0
}

func (t *Transport) totalTrafficСounting() {
	for key := range t.dataOld {
		if key.Alias == "Всего" {
			continue
		}
		t.Lock()
		// Подсчёт общего трафика
		//Создание пустой переменной типа KeyMapOfReports для осуществления действий с жёсткозаданным ключом
		keyTotal := KeyMapOfReports{}
		// Задаём заранее определенный ключ
		keyTotal.DateStr = key.DateStr
		keyTotal.Alias = "Всего"
		valueTotal := t.dataOld[keyTotal]
		value := t.dataOld[key]
		valueTotal.VolumePerDay += value.VolumePerDay
		valueTotal.Hits += value.Hits
		for index := range valueTotal.VolumePerHour {
			valueTotal.VolumePerHour[index] += value.VolumePerHour[index]
		}
		t.dataOld[keyTotal] = valueTotal
		t.Unlock()

	}
}

func (t *Transport) writeToChasheData() {
	t.Lock()
	t.dataCasheOld = MapOfReports{}
	t.dataCasheOld = t.dataOld
	t.Unlock()
}
