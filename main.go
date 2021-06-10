package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := newConfig()

	if !cfg.Debug {
		if err := CheckPIDFile(cfg.Pidfile); err != nil {
			log.Error(err)
			os.Exit(2)
		}
	}
	if err := writePID(cfg.Pidfile); err != nil {
		log.Error(err)
		os.Exit(2)
	}

	// TODO проверка на установленную программу
	// TODO если программа не установлена, то предложить установить её
	// TODO в случае согласия раскидать все файлы по папкам и установиться в systemd
	// TODO Еси это винда, то  скопировать файлы куда скажет пользователь и
	// TODO командами powershell провести запись в планировщике задач

	// TODO Сделать разделение на 3 части программы со своими параметрами, чтобы например... управление микротом было на запуск с ключом "-control", получение инфы по нетфлоу было с ключом "-flow", а просто статистика с ключом "-statistics"
	// TODO и кодовую базу разнести соответствующе (хотя бы попытаться)
	// TODO в последствии сделать так чтобы программа при запуске и проверки флагов запускала сама все три "ипостаси", а к примеру статистика следила за корректной работой остальных, в случа сбоя ребутила бы

	// TODO после "разделения" на три части сделать общение между частями по JSON или gRPC

	t := NewTransport(cfg)

	go t.Exit()
	go t.ReOpenLogAfterLogroatate()
	t.getAliasesArr(cfg)

	// Endless file parsing loop
	go func(cfg *Config) {
		t.runOnce(cfg)
		for {
			<-t.timerParse.C
			t.runOnce(cfg)
		}
	}(cfg)
	go t.pipeOutputToSquid(cfg)
	if !cfg.NoFlow {
		go t.readsStreamFromMT(cfg)
	}
	t.handleRequest(cfg)

}
