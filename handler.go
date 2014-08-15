package main

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
	"sync"
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
	hosts map[string]string
	dump  string
}

func NewBottDnsHandler(dump string) *BottDnsHandler {
	return &BottDnsHandler{
		&sync.Mutex{},
		make(map[string]string),
		dump,
	}
}

func (self *BottDnsHandler) Handle(w dns.ResponseWriter, req *dns.Msg) {
	q := req.Question[0]
	Q := Question{self.unFqdn(q.Name), dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass]}

	logger.Debug("Question:", Q.String())
	if self.isIPQuery(q) {
		if ip, ok := self.hosts[Q.qname]; ok {
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
