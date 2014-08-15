package main

import (
	"encoding/json"
	"github.com/bmizerany/pat"
	"net/http"
)

type Api struct {
	addr  string
	hosts *map[string][]string
}

func (self *Api) Detail(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Server", "Bott DNS server")
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.Encode(self.hosts)
}

func (self *Api) Serve() {
	m := pat.New()
	m.Get("/", http.HandlerFunc(self.Detail))

	http.Handle("/", m)
	logger.Info("Start api server", self.addr)
	err := http.ListenAndServe(self.addr, nil)
	if err != nil {
		logger.Debug("Start api failed", err)
		return
	}
}
