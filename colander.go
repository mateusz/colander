package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"text/tabwriter"

	"github.com/mateusz/sidereal/classifiers"
	"github.com/mateusz/sidereal/regimes"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
)

type config struct {
	verbose             bool
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
	logger = log.New(os.Stdout, "", 0)

	const (
		wantedLoadHelp          = "Total % of CPU utilisation to aim for when shaping the traffic"
		verboseHelp             = "Enable verbose output"
		backendHelp             = "Backend URI"
		listenPortHelp          = "Local HTTP listen port"
		bucketRollingWindowHelp = "Number of samples to keep in the bucket rolling window for both RPS and response time"
	)
	conf = &config{}
	flag.BoolVar(&conf.verbose, "verbose", true, verboseHelp)
	flag.IntVar(&conf.wantedLoad, "wanted-load", 120, wantedLoadHelp)
	flag.StringVar(&conf.backend, "backend", "http://localhost:80", backendHelp)
	flag.IntVar(&conf.listenPort, "listen-port", 8888, listenPortHelp)
	flag.IntVar(&conf.bucketRollingWindow, "bucket-rolling-window", 10, bucketRollingWindowHelp)
	flag.Parse()

	if conf.verbose {
		tw := tabwriter.NewWriter(os.Stdout, 24, 4, 1, ' ', tabwriter.AlignRight)
		fmt.Fprintf(tw, "Value\t   Option\f")
		fmt.Fprintf(tw, "%t\t - %s\f", conf.verbose, verboseHelp)
		fmt.Fprintf(tw, "%d\t - %s\f", conf.wantedLoad, wantedLoadHelp)
		fmt.Fprintf(tw, "%s\t - %s\f", conf.backend, backendHelp)
		fmt.Fprintf(tw, "%d\t - %s\f", conf.listenPort, listenPortHelp)
		fmt.Fprintf(tw, "%d\t - %s\f", conf.bucketRollingWindow, bucketRollingWindowHelp)
	}
}

func middleware(h http.Handler) http.Handler {
	crawler := classifiers.NewCrawler(logger)
	green := regimes.NewGreen(h, logger)
	prio1 := regimes.NewBucket(1, conf.bucketRollingWindow)
	prio2 := regimes.NewBucket(2, conf.bucketRollingWindow)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bucket := prio1
		if crawler.Belongs(r) {
			bucket = prio2
		}

		green.ServeHTTP(w, r, bucket)
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
