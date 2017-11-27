// Copyright 2016,2017 Yaacov Zamir <kobi.zamir@gmail.com>
// and other contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cli command line interface
package cli

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/MohawkTSDB/mohawk/backend"
	"github.com/MohawkTSDB/mohawk/backend/example"
	"github.com/MohawkTSDB/mohawk/backend/memory"
	"github.com/MohawkTSDB/mohawk/backend/mongo"
	"github.com/MohawkTSDB/mohawk/backend/sqlite"
	"github.com/MohawkTSDB/mohawk/middleware"
	"github.com/MohawkTSDB/mohawk/router"
)

// VER the server version
const VER = "0.21.4"

// AUTHOR the author name and Email
const AUTHOR = "Yaacov Zamir <kobi.zamir@gmail.com>"

// defaults
const defaultAPI = "0.21.0"
const defaultTLSKey = "server.key"
const defaultTLSCert = "server.pem"

// BackendName Mohawk active backend
var BackendName string

// RootCmd Mohawk root cli Command
var RootCmd = &cobra.Command{
	Use: "mohawk",
	Long: fmt.Sprintf(`Mohawk is a metric data storage engine.

Mohawk is a metric data storage engine that uses a plugin architecture for data
storage and a simple REST API as the primary interface.

Version:
  %s

Author:
  %s`, VER, AUTHOR),
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	// Flag definition
	RootCmd.Flags().StringP("backend", "b", "memory", "the backend plugin to use")
	RootCmd.Flags().String("token", "", "authorization token")
	RootCmd.Flags().String("media", "./mohawk-webui", "path to media files")
	RootCmd.Flags().String("key", defaultTLSKey, "path to TLS key file")
	RootCmd.Flags().String("cert", defaultTLSCert, "path to TLS cert file")
	RootCmd.Flags().String("options", "", "specific backend options [e.g. db-dirname, db-url]")
	RootCmd.Flags().IntP("port", "p", 8080, "server port")
	RootCmd.Flags().BoolP("tls", "t", false, "use TLS server")
	RootCmd.Flags().BoolP("gzip", "g", false, "use gzip encoding")
	RootCmd.Flags().BoolP("verbose", "V", false, "more debug output")
	RootCmd.Flags().BoolP("version", "v", false, "display mohawk version number")
	RootCmd.Flags().StringP("config", "c", "", "config file (default is $HOME/.cobra.yaml)")

	// Viper Binding
	viper.BindPFlag("backend", RootCmd.Flags().Lookup("backend"))
	viper.BindPFlag("token", RootCmd.Flags().Lookup("token"))
	viper.BindPFlag("media", RootCmd.Flags().Lookup("media"))
	viper.BindPFlag("key", RootCmd.Flags().Lookup("key"))
	viper.BindPFlag("cert", RootCmd.Flags().Lookup("cert"))
	viper.BindPFlag("options", RootCmd.Flags().Lookup("options"))
	viper.BindPFlag("port", RootCmd.Flags().Lookup("port"))
	viper.BindPFlag("tls", RootCmd.Flags().Lookup("tls"))
	viper.BindPFlag("gzip", RootCmd.Flags().Lookup("gzip"))
	viper.BindPFlag("verbose", RootCmd.Flags().Lookup("verbose"))
	viper.BindPFlag("version", RootCmd.Flags().Lookup("version"))
	viper.BindPFlag("config", RootCmd.Flags().Lookup("config"))
}

func initConfig() {
	if viper.GetString("config") != "" {
		viper.SetConfigFile(viper.GetString("config"))
		if err := viper.ReadInConfig(); err != nil {
			fmt.Println("Error reading config file:", err)
		}
	}
}

// GetStatus return a json status struct
func GetStatus(w http.ResponseWriter, r *http.Request, argv map[string]string) {
	resTemplate := `{"MetricsService":"STARTED","Implementation-Version":"%s","MohawkVersion":"%s","MohawkBackend":"%s"}`
	res := fmt.Sprintf(resTemplate, defaultAPI, VER, BackendName)

	w.WriteHeader(200)
	fmt.Fprintln(w, res)
}

func serve() error {
	var db backend.Backend
	var middlewareList []middleware.MiddleWare

	var backendQuery = viper.GetString("backend")
	var optionsQuery = viper.GetString("options")
	var verbose = viper.GetBool("verbose")
	var media = viper.GetString("media")
	var tls = viper.GetBool("tls")
	var gzip = viper.GetBool("gzip")
	var token = viper.GetString("token")
	var port = viper.GetInt("port")
	var cert = viper.GetString("cert")
	var key = viper.GetString("key")

	if viper.GetBool("version") {
		fmt.Printf("Mohawk version: %s\n\n", VER)
		return nil
	}

	// Create and init the backend
	switch backendQuery {
	case "sqlite":
		db = &sqlite.Backend{}
	case "memory":
		db = &memory.Backend{}
	case "mongo":
		db = &mongo.Backend{}
	case "example":
		db = &example.Backend{}
	default:
		log.Fatal("Can't find backend:", backendQuery)
	}

	// parse options
	if options, err := url.ParseQuery(optionsQuery); err == nil {
		db.Open(options)
	} else {
		log.Fatal("Can't parse opetions:", optionsQuery)
	}

	// set global variables
	BackendName = db.Name()

	// h common variables to be used for the backend Handler functions
	// backend the backend to use for metrics source
	h := backend.Handler{
		Verbose: verbose,
		Backend: db,
	}

	// Create the routers
	// Requests not handled by the routers will be forworded to BadRequest Handler
	rRoot := router.Router{
		Prefix: "/hawkular/metrics/",
	}
	// Root Routing table
	rRoot.Add("GET", "status", GetStatus)
	rRoot.Add("GET", "tenants", h.GetTenants)
	rRoot.Add("GET", "metrics", h.GetMetrics)

	// Metrics Routing tables
	rGauges := router.Router{
		Prefix: "/hawkular/metrics/gauges/",
	}
	rGauges.Add("GET", ":id/raw", h.GetData)
	rGauges.Add("GET", ":id/stats", h.GetData)
	rGauges.Add("POST", "raw", h.PostData)
	rGauges.Add("POST", "raw/query", h.PostQuery)
	rGauges.Add("PUT", "tags", h.PutMultiTags)
	rGauges.Add("PUT", ":id/tags", h.PutTags)
	rGauges.Add("DELETE", ":id/raw", h.DeleteData)
	rGauges.Add("DELETE", ":id/tags/:tags", h.DeleteTags)

	// deprecated
	rGauges.Add("GET", ":id/data", h.GetData)
	rGauges.Add("POST", "data", h.PostData)
	rGauges.Add("POST", "stats/query", h.PostQuery)

	rCounters := router.Router{
		Prefix: "/hawkular/metrics/counters/",
	}
	rCounters.Add("GET", ":id/raw", h.GetData)
	rCounters.Add("GET", ":id/stats", h.GetData)
	rCounters.Add("POST", "raw", h.PostData)
	rCounters.Add("POST", "raw/query", h.PostQuery)
	rCounters.Add("PUT", ":id/tags", h.PutTags)

	// deprecated
	rCounters.Add("GET", ":id/data", h.GetData)
	rCounters.Add("POST", "data", h.PostData)
	rCounters.Add("POST", "stats/query", h.PostQuery)

	rAvailability := router.Router{
		Prefix: "/hawkular/metrics/availability/",
	}
	rAvailability.Add("GET", ":id/raw", h.GetData)
	rAvailability.Add("GET", ":id/stats", h.GetData)

	// Create the middlewares
	// logger a logging middleware
	logger := middleware.Logger{
		Verbose: verbose,
	}

	// static a file server middleware
	static := middleware.Static{
		Verbose:   verbose,
		MediaPath: media,
	}

	// authorization middleware
	authorization := middleware.Authorization{
		Verbose:         verbose,
		UseToken:        token != "",
		PublicPathRegex: "^/hawkular/metrics/status$",
		Token:           token,
	}

	// headers a headers middleware
	headers := middleware.Headers{
		Verbose: verbose,
	}

	// gzipper a gzip encoding middleware
	gzipper := middleware.GZipper{
		UseGzip: gzip,
		Verbose: verbose,
	}

	// badrequest a BadRequest middleware
	badrequest := middleware.BadRequest{
		Verbose: verbose,
	}

	// concat middlewars and routes (first logger until rRoot) with a fallback to BadRequest
	middlewareList = []middleware.MiddleWare{
		&logger,
		&authorization,
		&headers,
		&gzipper,
		&rGauges,
		&rCounters,
		&rAvailability,
		&rRoot,
		&static,
		&badrequest,
	}
	middleware.Append(middlewareList)

	// Run the server
	srv := &http.Server{
		Addr:           fmt.Sprintf("0.0.0.0:%d", port),
		Handler:        logger,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if tls {
		log.Printf("Start server, listen on https://%+v", srv.Addr)
		log.Fatal(srv.ListenAndServeTLS(cert, key))
	} else {
		log.Printf("Start server, listen on http://%+v", srv.Addr)
		log.Fatal(srv.ListenAndServe())
	}

	return nil
}