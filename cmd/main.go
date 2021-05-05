package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"

	"github.com/CXinsect/gocode/config"
	"github.com/CXinsect/gocode/prom"
)

func main() {

	serverIP := kingpin.Flag("listen-address", "The Server address to listen on for http").Default(":8888").String()
	configFile := kingpin.Flag("config.file", "The path for config file").Default("config.yaml").File()
	labelIP := kingpin.Flag("label-ip", "The unique ip for labels of metrics").String()
	version := kingpin.Flag("version", "The Info of exporter").Default("0.0.0").String()
	kingpin.Parse()

	config := config.Config{}
	err := yaml.NewDecoder(*configFile).Decode(&config)
	if err != nil {
		log.Fatalf("The config file obtained error")
	}

	exporter := prom.New(config)
	err = exporter.Attach()
	if err != nil {
		log.Fatalf("exporter attach failed %s", err)
	}

	err = prometheus.Register(exporter)

	if err != nil {
		log.Fatalf("The prometheus has an error in registering exporter")
	}
	http.Handle("/metrics", promhttp.Handler())

	log.Printf("Listen on Address %s with exporter version %s with label-ip %s", *serverIP, *version, *labelIP)

	err = http.ListenAndServe(*serverIP, nil)

	if err != nil {
		log.Fatalf("http server started failed %s", err)
	}
}
