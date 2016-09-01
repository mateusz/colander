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
	"time"

	"golang.org/x/net/context"
	"golang.org/x/time/rate"

	"github.com/mateusz/sidereal/classifiers"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
)

type config struct {
	verbose    bool
	wantedLoad int
	backend    string
	listenPort int
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
		wantedLoadHelp = "Total % of CPU utilisation to aim for when shaping the traffic"
		verboseHelp    = "Enable verbose output"
		backendHelp    = "Backend URI"
		listenPortHelp = "Local HTTP listen port"
	)
	conf = &config{}
	flag.BoolVar(&conf.verbose, "verbose", true, verboseHelp)
	flag.IntVar(&conf.wantedLoad, "wanted-load", 120, wantedLoadHelp)
	flag.StringVar(&conf.backend, "backend", "http://localhost:80", backendHelp)
	flag.IntVar(&conf.listenPort, "listen-port", 8888, listenPortHelp)
	flag.Parse()

	if conf.verbose {
		tw := tabwriter.NewWriter(os.Stdout, 24, 4, 1, ' ', tabwriter.AlignRight)
		fmt.Fprintf(tw, "Value\t   Option\f")
		fmt.Fprintf(tw, "%t\t - %s\f", conf.verbose, verboseHelp)
		fmt.Fprintf(tw, "%d\t - %s\f", conf.wantedLoad, wantedLoadHelp)
		fmt.Fprintf(tw, "%s\t - %s\f", conf.backend, backendHelp)
		fmt.Fprintf(tw, "%d\t - %s\f", conf.listenPort, listenPortHelp)
	}
}

func middleware(h http.Handler) http.Handler {
	c := classifiers.NewCrawler(logger)
	// 1.0rps, 10 burst (bucket depth)
	crawlerBucket := rate.NewLimiter(1.0, 10)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c.Belongs(r) {
			// Max wait 10s (i.e. permit if the queue length, considering rps and burst, is less than 10s,
			// Otherwise fail outright.
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			err := crawlerBucket.Wait(ctx)
			if err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
		}
		h.ServeHTTP(w, r)
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
