package main

import (
	"fmt"
	"github.com/miekg/dns"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

type Question struct {
	qname  string
	qtype  string
	qclass string
}

func (self *Question) String() string {
	return fmt.Sprintf("%s %s %s", self.qname, self.qclass, self.qtype)
}

type BottDnsHandler struct {
	mu    *sync.Mutex
	hosts map[string][]string
	dump  string
}

func NewBottDnsHandler(dump string) *BottDnsHandler {
	h := &BottDnsHandler{
		mu:   &sync.Mutex{},
		dump: dump,
	}
	h.Load()
	return h
}

func (self *BottDnsHandler) Load() {
	self.hosts = make(map[string][]string)
	if _, err := os.Stat(self.dump); err == nil {
		b, err := ioutil.ReadFile(self.dump)
		if err != nil {
			logger.Info("Read dump file failed", self.dump)
		}
		if err := yaml.Unmarshal(b, &self.hosts); err != nil {
			logger.Info("Load dump file failed", self.dump)
		}
	}
	logger.Debug(self.hosts)
}

func (self *BottDnsHandler) Dump() {
	self.mu.Lock()
	defer self.mu.Unlock()
	for domain, ips := range self.hosts {
		if len(ips) == 0 {
			delete(self.hosts, domain)
		}
	}
	b, err := yaml.Marshal(self.hosts)
	if err != nil {
		logger.Info("Dump hosts failed")
		return
	}
	if err := ioutil.WriteFile(self.dump, b, 0755); err != nil {
		logger.Info("Save dump failed")
		return
	}
}

func (self *BottDnsHandler) Responder(w dns.ResponseWriter, req *dns.Msg, ips []string) {
	q := req.Question[0]
	m := new(dns.Msg)
	m.SetReply(req)
	rr_header := dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600}
	for _, ip := range ips {
		a := &dns.A{rr_header, net.ParseIP(ip)}
		m.Answer = append(m.Answer, a)
	}
	w.WriteMsg(m)
	return
}

func (self *BottDnsHandler) Handle(w dns.ResponseWriter, req *dns.Msg) {
	rand.Seed(time.Now().UnixNano())
	q := req.Question[0]
	Q := Question{self.unFqdn(q.Name), dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass]}

	logger.Debug("Question:", Q.String())
	if ok := net.ParseIP(Q.qname); ok != nil {
		self.Responder(w, req, []string{Q.qname})
		return
	}
	if self.isIPQuery(q) {
		if ips, ok := self.hosts[Q.qname]; ok {
			self.Responder(w, req, ips)
			logger.Debug(Q.qname, "found in hosts")
			return
		}
	}
	logger.Debug(Q.qname, "not found")
	dns.HandleFailed(w, req)
}

func (self *BottDnsHandler) isIPQuery(q dns.Question) bool {
	return q.Qtype == dns.TypeA && q.Qclass == dns.ClassINET
}

func (self *BottDnsHandler) unFqdn(s string) string {
	if dns.IsFqdn(s) {
		return s[:len(s)-1]
	}
	return s
}
