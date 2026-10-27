// Serve browser checks and proxy Catalog through the same local origin.
package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:8082", "local check server address")
	catalog := flag.String("catalog", "http://127.0.0.1:8081/md3", "running Catalog base URL")
	profilePIDs := flag.String("profile-pids", "", "optional comma-separated Linux renderer/GPU PIDs for CPU measurements")
	flag.Parse()
	target, err := url.Parse(*catalog)
	if err != nil || target.Host == "" || (target.Scheme != "http" && target.Scheme != "https") || target.RawQuery != "" || target.Fragment != "" {
		log.Fatal("catalog must be an HTTP(S) base URL without a query or fragment")
	}
	counters, err := cpuProfile(*profilePIDs)
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /profile/cpu", counters)
	proxy := httputil.NewSingleHostReverseProxy(target)
	mux.Handle("/md3/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/md3")
		proxy.ServeHTTP(w, r)
	}))
	for _, route := range []string{"/components", "/components/", "/styles", "/styles/", "/static/", "/themes/", "/theme.css", "/changes", "/for-agents", "/sitemap.xml"} {
		mux.Handle(route, proxy)
	}
	files := http.FileServer(http.Dir("forge/conformance"))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		files.ServeHTTP(w, r)
	}))
	server := &http.Server{Addr: *listen, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("Component checks: http://%s/motion.html", *listen)
	log.Fatal(server.ListenAndServe())
}
