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

// Package handler
package handler

import (
	"net/http"
)

// Headers middleware that will log http requests
type Headers struct {
	Verbose bool
	next    http.Handler
}

// SetNext set next http serve func
func (l *Headers) SetNext(h http.Handler) {
	l.next = h
}

// ServeHTTP http serve func
func (l Headers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "authorization,content-type,hawkular-tenant")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT")

	l.next.ServeHTTP(w, r)
}
