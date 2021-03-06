/*
Copyright 2020 The Flux CD contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testserver

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
)

func NewTempHTTPServer() (*HTTP, error) {
	tmpDir, err := ioutil.TempDir("", "http-test-")
	if err != nil {
		return nil, err
	}
	srv := NewHTTPServer(tmpDir)
	return srv, nil
}

func NewHTTPServer(docroot string) *HTTP {
	root, err := filepath.Abs(docroot)
	if err != nil {
		panic(err)
	}
	return &HTTP{
		docroot: root,
	}
}

type HTTP struct {
	docroot    string
	middleware func(http.Handler) http.Handler
	server     *httptest.Server
}

func (s *HTTP) WithMiddleware(m func(handler http.Handler) http.Handler) *HTTP {
	s.middleware = m
	return s
}

func (s *HTTP) Start() {
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := http.FileServer(http.Dir(s.docroot))
		if s.middleware != nil {
			s.middleware(handler).ServeHTTP(w, r)
			return
		}
		handler.ServeHTTP(w, r)
	}))
}

func (s *HTTP) StartTLS(cert, key, ca []byte) error {
	s.server = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := http.FileServer(http.Dir(s.docroot))
		if s.middleware != nil {
			s.middleware(handler).ServeHTTP(w, r)
			return
		}
		handler.ServeHTTP(w, r)
	}))

	config := tls.Config{}

	keyPair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return err
	}
	config.Certificates = []tls.Certificate{keyPair}

	cp := x509.NewCertPool()
	cp.AppendCertsFromPEM(ca)
	config.RootCAs = cp

	config.BuildNameToCertificate()
	config.ServerName = "example.com"
	s.server.TLS = &config

	s.server.StartTLS()
	return nil
}

func (s *HTTP) Stop() {
	if s.server != nil {
		s.server.Close()
	}
}

func (s *HTTP) Root() string {
	return s.docroot
}

func (s *HTTP) URL() string {
	if s.server != nil {
		return s.server.URL
	}
	return ""
}
