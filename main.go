package main

func main() {
	cfg := newConfig()
	// TODO проерка на запущенный экземпляр

	transport := NewTransport(cfg)

	go transport.Exit()
	go transport.ReOpenLogAfterLogroatate()
	transport.getAllAliases(cfg)
	go transport.loopGetDataFromMT()
	go transport.loopParse(cfg)
	go transport.pipeOutputToStdoutForSquid(cfg)
	transport.handleRequest()
	transport.readsStreamFromMT(cfg)
}
