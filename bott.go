package main

import (
	"fmt"
	"github.com/miekg/dns"
	"net/url"
	"os"
	"syscall"
	"time"
)

type Bott struct {
	addr     string
	rTimeout int
	wTimeout int
}

func (self *Bott) Serve(dump string, c chan os.Signal) {
	u, err := url.Parse(self.addr)
	if err != nil {
		logger.Assert(err, "url")
	}

	handler := NewBottDnsHandler(dump)
	defer handler.Dump()

	dnsHandler := dns.NewServeMux()
	dnsHandler.HandleFunc(".", handler.Handle)

	logger.Debug(u.Scheme, u.Host)

	server := &dns.Server{
		Addr:         u.Host,
		Net:          u.Scheme,
		Handler:      dnsHandler,
		ReadTimeout:  time.Duration(self.rTimeout) * time.Second,
		WriteTimeout: time.Duration(self.wTimeout) * time.Second,
	}

	if u.Scheme == "udp" {
		server.UDPSize = 65535
	}

	go func() {
		logger.Info("Start on", self.addr)
		err := server.ListenAndServe()
		if err != nil {
			logger.Assert(err, fmt.Sprintf("Start failed on %s", self.addr))
		}
	}()

	for {
		s := <-c
		logger.Info("Catch", s)
		switch s {
		case syscall.SIGHUP:
			handler.Dump()
		default:
			return
		}
	}
}
