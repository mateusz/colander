package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"text/tabwriter"

	log "github.com/Sirupsen/logrus"
	"github.com/mateusz/colander/shaper"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
)

type config struct {
	debug               bool
	wantedLoad          int
	backend             string
	listenPort          int
	bucketRollingWindow int
}

var logger *log.Logger
var conf *config

func utilisation() (int, error) {
	load, err := load.Avg()
	if err != nil {
		return -1, err
	}
	count, err := cpu.Counts(true)
	if err != nil {
		return -1, err
	}

	return int(load.Load1 * 100 / float64(count)), nil

}

func init() {
	log.SetOutput(os.Stderr)
	log.SetLevel(log.WarnLevel)

	const (
		wantedLoadHelp          = "Total % of CPU utilisation to aim for when shaping the traffic"
		debugHelp               = "Enable debug output"
		backendHelp             = "Backend URI"
		listenPortHelp          = "Local HTTP listen port"
		bucketRollingWindowHelp = "Number of samples to keep in the bucket rolling window for both RPS and response time"
	)
	conf = &config{}
	flag.BoolVar(&conf.debug, "debug", false, debugHelp)
	flag.IntVar(&conf.wantedLoad, "wanted-load", 120, wantedLoadHelp)
	flag.StringVar(&conf.backend, "backend", "http://localhost:80", backendHelp)
	flag.IntVar(&conf.listenPort, "listen-port", 8888, listenPortHelp)
	flag.IntVar(&conf.bucketRollingWindow, "bucket-rolling-window", 10, bucketRollingWindowHelp)
	flag.Parse()

	if conf.debug {
		tw := tabwriter.NewWriter(os.Stdout, 24, 4, 1, ' ', tabwriter.AlignRight)
		fmt.Fprintf(tw, "Value\t   Option\f")
		fmt.Fprintf(tw, "%t\t - %s\f", conf.debug, debugHelp)
		fmt.Fprintf(tw, "%d\t - %s\f", conf.wantedLoad, wantedLoadHelp)
		fmt.Fprintf(tw, "%s\t - %s\f", conf.backend, backendHelp)
		fmt.Fprintf(tw, "%d\t - %s\f", conf.listenPort, listenPortHelp)
		fmt.Fprintf(tw, "%d\t - %s\f", conf.bucketRollingWindow, bucketRollingWindowHelp)

		log.SetLevel(log.DebugLevel)
	}
}

func middleware(h http.Handler) http.Handler {
	classifier := shaper.ClassifierFunc(func(r *http.Request) shaper.Class {
		return shaper.Class(1)
	})
	green := shaper.NewGreen(h)
	shaper := shaper.New(classifier, green)
	shaper.BucketRollingWindow = conf.bucketRollingWindow
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shaper.ServeHTTP(w, r)
	})
}

func listen() {
	url, err := url.Parse(conf.backend)
	if err != nil {
		logger.Fatalln(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	handler := middleware(proxy)
	http.ListenAndServe(fmt.Sprintf(":%d", conf.listenPort), handler)
}

func main() {
	listen()
}
