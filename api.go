package main

import (
	"encoding/json"
	"github.com/bmizerany/pat"
	"net/http"
)

type Message struct {
	Ip []string
}

type Api struct {
	addr  string
	hosts map[string][]string
}

func (self *Api) header(w http.ResponseWriter) {
	w.Header().Set("Server", "Bott DNS server")
	w.Header().Set("Content-Type", "application/json")
}

func (self *Api) Detail(w http.ResponseWriter, req *http.Request) {
	self.header(w)
	encoder := json.NewEncoder(w)
	encoder.Encode(self.hosts)
}

func (self *Api) Add(w http.ResponseWriter, req *http.Request) {
	host := req.URL.Query().Get(":host")
	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(req.Body)
	data := Message{}
	err := decoder.Decode(&data)
	if err != nil {
		logger.Debug(err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	self.header(w)
	logger.Debug(host, data)
	ips, ok := self.hosts[host]
	if ok {
		vetrx := make(map[string]struct{})
		for _, ip := range ips {
			vetrx[ip] = struct{}{}
		}
		for _, ip := range data.Ip {
			if _, ok := vetrx[ip]; ok {
				continue
			}
			self.hosts[host] = append(self.hosts[host], ip)
		}
	} else {
		self.hosts[host] = make([]string, len(data.Ip))
		copy(self.hosts[host], data.Ip)
	}
	encoder.Encode(self.hosts)
}

func (self *Api) Delete(w http.ResponseWriter, req *http.Request) {
	self.header(w)
}

func (self *Api) Serve() {
	m := pat.New()
	m.Get("/", http.HandlerFunc(self.Detail))
	m.Put("/host/:host", http.HandlerFunc(self.Add))
	m.Del("/host/:host", http.HandlerFunc(self.Delete))

	http.Handle("/", m)
	logger.Info("Start api server", self.addr)
	err := http.ListenAndServe(self.addr, nil)
	if err != nil {
		logger.Debug("Start api failed", err)
		return
	}
}
