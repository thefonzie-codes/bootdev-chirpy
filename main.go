package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
    fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { 
        cfg.fileserverHits.Add(1)
        next.ServeHTTP(w, r)
    })
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, _ *http.Request) {
    hits := cfg.fileserverHits.Load()
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Hits: %d", hits)
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, _ *http.Request) {
    cfg.fileserverHits.Store((0))
    w.WriteHeader(http.StatusOK)
}

func main () {
    mux := http.NewServeMux()

    s := &http.Server {
        Addr: ":8080",
        Handler: mux,
    }

    apiCfg := apiConfig{}

    mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
    mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Add("Content-Type", "text/plain; charset=utf-8")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    mux.HandleFunc("GET /metrics", apiCfg.metricsHandler)
    mux.HandleFunc("POST /reset", apiCfg.resetHandler)

    log.Fatal(s.ListenAndServe())
}
