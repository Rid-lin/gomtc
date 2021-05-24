package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

func (t *Transport) getAliasesArr(cfg *Config) {
	aliases := make(map[string][]string)
	path := path.Join(cfg.ConfigPath, "realname.cfg")
	buf, err := os.Open(path)
	if err != nil {
		log.Error(err)
		return
	}
	defer func() {
		if err = buf.Close(); err != nil {
			log.Error(err)
		}
	}()
	snl := bufio.NewScanner(buf)
	for snl.Scan() {
		line := snl.Text()
		lineArray := strings.Split(line, " ")
		if len(lineArray) <= 1 {
			continue
		}
		key := lineArray[0]
		value := lineArray[1:]
		if test, ok := aliases[key]; ok {
			aliases[key] = append(test, fmt.Sprint(value))
		} else {
			aliases[key] = []string{fmt.Sprint(value)}
		}
	}
	err = snl.Err()
	if err != nil {
		log.Error(err)
		return
	}
	t.Lock()
	t.AliasesStrArr = aliases
	t.Unlock()
	log.Trace(aliases)
}
