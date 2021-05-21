package main

func main() {
	cfg := newConfig()
	// TODO проверка на запущенный экземпляр
	// TODO проверка на установленную программу
	// TODO если программа не установлена, то предложить установить её
	// TODO в случае согласия раскидать все файлы по папкам и установиться в systemd
	// TODO Еси это винда, то  скопировать файлы куда скажет пользователь и
	// TODO командами powershell провести запись в планировщике задач

	transport := NewTransport(cfg)

	go transport.Exit()
	go transport.ReOpenLogAfterLogroatate()
	transport.getAllAliases(cfg)
	go transport.loopGetDataFromMT()
	go transport.loopParse(cfg)
	go transport.pipeOutputToSquid(cfg)
	if !cfg.NoFlow {
		go transport.readsStreamFromMT(cfg)
	}
	transport.handleRequest(cfg)

}
