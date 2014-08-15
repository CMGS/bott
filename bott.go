package main

import (
	"fmt"
	"github.com/miekg/dns"
	"net/url"
	"time"
)

type Bott struct {
	addr     string
	rTimeout int
	wTimeout int
}

func (self *Bott) Serve(dump string) {
	u, err := url.Parse(self.addr)
	if err != nil {
		logger.Assert(err, "url")
	}

	Handler := NewBottDnsHandler(dump)

	dnsHandler := dns.NewServeMux()
	dnsHandler.HandleFunc(".", Handler.Handle)

	logger.Debug(u.Scheme, u.Host)

	server := &dns.Server{
		Addr:         u.Host,
		Net:          u.Scheme,
		Handler:      dnsHandler,
		ReadTimeout:  time.Duration(self.rTimeout) * time.Second,
		WriteTimeout: time.Duration(self.wTimeout) * time.Second,
	}

	go func() {
		logger.Info("Start on", self.addr)
		err := server.ListenAndServe()
		if err != nil {
			logger.Assert(err, fmt.Sprintf("Start failed on %s", self.addr))
		}
	}()
}
