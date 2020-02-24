package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage", os.Args[0], "<config_file>")
		os.Exit(1)
	}

	config := NewConfig(os.Args[1])
	client := &http.Client{
		Timeout: config.Timeout,
	}
	collector := NewCollector(*config.InfoURL, client)

	prometheus.MustRegister(&collector)
	http.Handle("/metrics", promhttp.Handler())
	log.Println(http.ListenAndServe(config.Listen, nil))
}
