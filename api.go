package main

import (
	"encoding/json"
	"github.com/bmizerany/pat"
	"net/http"
	"sync"
)

type Message struct {
	Ip []string
}

type Api struct {
	addr  string
	hosts map[string][]string
	mu    *sync.Mutex
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

func makeVetrx(ips []string) map[string]struct{} {
	vetrx := make(map[string]struct{}, len(ips))
	for _, ip := range ips {
		vetrx[ip] = struct{}{}
	}
	logger.Debug("vetrx", vetrx)
	return vetrx
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

	self.mu.Lock()
	defer self.mu.Unlock()

	if ok {
		vetrx := makeVetrx(ips)
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

	self.mu.Lock()
	defer self.mu.Unlock()

	if !ok {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	vetrx := makeVetrx(ips)
	for _, ip := range data.Ip {
		if _, ok := vetrx[ip]; !ok {
			continue
		}
		delete(vetrx, ip)
	}

	if len(vetrx) == 0 {
		delete(self.hosts, host)
	} else {
		new_ips := make([]string, 0, len(vetrx))
		for ip, _ := range vetrx {
			new_ips = append(new_ips, ip)
		}
		self.hosts[host] = new_ips
	}
	encoder.Encode(self.hosts)
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
