package main

func main() {
	cfg := newConfig()
	// TODO проерка на запущенный экземпляр
	// TODO проерка на установленную программу
	// TODO если программа не установлена, то предложить установить её
	// TODO в случае согласия раскидать все файлы по папкам и установиться в systemd

	transport := NewTransport(cfg)

	go transport.Exit()
	go transport.ReOpenLogAfterLogroatate()
	transport.getAllAliases(cfg)
	go transport.loopGetDataFromMT()
	go transport.loopParse(cfg)
	go transport.pipeOutputToSquid(cfg)
	transport.handleRequest(cfg)
	transport.readsStreamFromMT(cfg)
}
