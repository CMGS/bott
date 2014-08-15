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
		mu:    &sync.Mutex{},
		dump:  dump,
		hosts: make(map[string][]string),
	}
	if _, err := os.Stat(dump); err == nil {
		b, err := ioutil.ReadFile(dump)
		if err != nil {
			logger.Info("Read dump file failed", dump)
		}
		if err := yaml.Unmarshal(b, &h.hosts); err != nil {
			logger.Info("Load dump file failed", dump)
		}
	}
	logger.Debug(h.hosts)
	return h
}

func (self *BottDnsHandler) Dump() {
	self.mu.Lock()
	defer self.mu.Unlock()
}

func (self *BottDnsHandler) Handle(w dns.ResponseWriter, req *dns.Msg) {
	rand.Seed(time.Now().UnixNano())
	q := req.Question[0]
	Q := Question{self.unFqdn(q.Name), dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass]}

	logger.Debug("Question:", Q.String())
	if self.isIPQuery(q) {
		if ips, ok := self.hosts[Q.qname]; ok {
			ip := ips[rand.Intn(len(ips))]
			m := new(dns.Msg)
			m.SetReply(req)
			rr_header := dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET}
			a := &dns.A{rr_header, net.ParseIP(ip)}
			m.Answer = append(m.Answer, a)
			w.WriteMsg(m)
			logger.Debug(Q.qname, "found in hosts", Q.qname)
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
