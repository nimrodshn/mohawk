package main

import (
	"log"
	"net/http"
	"time"

	"github.com/yaacov/mohawk/backends"
	"github.com/yaacov/mohawk/router"
)

func main() {
	db := backend.Random{}
	db.Open()

	h := Handler{
		backend: db,
		version: "0.21.0",
	}
	r := router.Router{
		Prefix:           "/hawkular/metrics/",
		HandleBadRequest: h.BadRequest,
	}

	r.Add("GET", "status", h.GetStatus)
	r.Add("GET", "metrics", h.GetMetrics)
	r.Add("GET", "gauges/:id/raw", h.GetData)
	r.Add("GET", "counters/:id/raw", h.GetData)

	srv := &http.Server{
		Addr:           "0.0.0.0:8443",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("Start server, listen on https://%+v", srv.Addr)
	log.Fatal(srv.ListenAndServeTLS("server.pem", "server.key"))
}
