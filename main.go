package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var rTimeout, wTimeout int
	var addr, dump string

	flag.BoolVar(&logger.Mode, "DEBUG", false, "enable debug")
	flag.StringVar(&addr, "addr", "udp://127.0.0.1:9999", "server address")
	flag.IntVar(&rTimeout, "read-timeout", 5, "server read timeout")
	flag.IntVar(&wTimeout, "write-timeout", 5, "server write timeout")
	flag.StringVar(&dump, "dump", "/tmp/dump.yaml", "dump file path")
	flag.Parse()

	bott := Bott{
		addr, rTimeout, wTimeout,
	}
	bott.Serve(dump)

	var c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGHUP)
	for {
		s := <-c
		logger.Info("Catch", s)
		switch s {
		case syscall.SIGHUP:
			continue
		default:
			os.Exit(0)
		}
	}
}
