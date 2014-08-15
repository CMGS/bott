package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func pid(path string) {
	if err := ioutil.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0755); err != nil {
		logger.Info("Save pid file failed", err)
	}
}

func main() {
	var rTimeout, wTimeout int
	var addr, dump, pidFile, api string

	flag.BoolVar(&logger.Mode, "DEBUG", false, "enable debug")
	flag.StringVar(&addr, "addr", "udp://127.0.0.1:9999", "server address")
	flag.StringVar(&api, "api", "127.0.0.1:8080", "api address")
	flag.IntVar(&rTimeout, "read-timeout", 5, "server read timeout")
	flag.IntVar(&wTimeout, "write-timeout", 5, "server write timeout")
	flag.StringVar(&dump, "dump", "/tmp/dump.yaml", "dump file path")
	flag.StringVar(&pidFile, "pidfile", "/var/run/bott.pid", "pid file")

	flag.Parse()

	var c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGHUP)

	pid(pidFile)
	defer os.Remove(pidFile)

	bott := Bott{
		addr, api, rTimeout, wTimeout,
	}
	bott.Serve(dump, c)
}
