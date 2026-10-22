// Serve browser checks and proxy Catalog through the same local origin.
package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:8082", "local check server address")
	catalog := flag.String("catalog", "http://127.0.0.1:8081", "running Catalog origin")
	profilePIDs := flag.String("profile-pids", "", "optional comma-separated Linux renderer/GPU PIDs for CPU measurements")
	flag.Parse()
	target, err := url.Parse(*catalog)
	if err != nil || target.Host == "" || (target.Scheme != "http" && target.Scheme != "https") || target.Path != "" {
		log.Fatal("catalog must be an HTTP(S) origin without a path")
	}
	counters, err := cpuProfile(*profilePIDs)
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /profile/cpu", counters)
	mux.Handle("/md3/", httputil.NewSingleHostReverseProxy(target))
	files := http.FileServer(http.Dir("forge/conformance"))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		files.ServeHTTP(w, r)
	}))
	server := &http.Server{Addr: *listen, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("Component checks: http://%s/motion.html", *listen)
	log.Fatal(server.ListenAndServe())
}
