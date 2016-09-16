package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"text/tabwriter"

	"golang.org/x/time/rate"

	log "github.com/Sirupsen/logrus"
	"github.com/mateusz/colander/deciders"
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
	loadThreshold       float64
}

var logger *log.Logger
var conf *config

func utilisation() (float64, error) {
	load, err := load.Avg()
	if err != nil {
		return -1, err
	}
	count, err := cpu.Counts(true)
	if err != nil {
		return -1, err
	}

	return load.Load1 / float64(count), nil

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
		loadThresholdHelp       = "Load threshold for enabling the limiter (e.g. on 4 core machine, 1.0 equals 4.0 1-minute load average"
	)
	conf = &config{}
	flag.BoolVar(&conf.debug, "debug", false, debugHelp)
	flag.IntVar(&conf.wantedLoad, "wanted-load", 120, wantedLoadHelp)
	flag.StringVar(&conf.backend, "backend", "http://localhost:80", backendHelp)
	flag.IntVar(&conf.listenPort, "listen-port", 8888, listenPortHelp)
	flag.IntVar(&conf.bucketRollingWindow, "bucket-rolling-window", 10, bucketRollingWindowHelp)
	flag.Float64Var(&conf.loadThreshold, "load-threshold", 1.2, loadThresholdHelp)
	flag.Parse()

	if conf.debug {
		tw := tabwriter.NewWriter(os.Stdout, 24, 4, 1, ' ', tabwriter.AlignRight)
		fmt.Fprintf(tw, "Value\t   Option\f")
		fmt.Fprintf(tw, "%t\t - %s\f", conf.debug, debugHelp)
		fmt.Fprintf(tw, "%d\t - %s\f", conf.wantedLoad, wantedLoadHelp)
		fmt.Fprintf(tw, "%s\t - %s\f", conf.backend, backendHelp)
		fmt.Fprintf(tw, "%d\t - %s\f", conf.listenPort, listenPortHelp)
		fmt.Fprintf(tw, "%d\t - %s\f", conf.bucketRollingWindow, bucketRollingWindowHelp)
		fmt.Fprintf(tw, "%.2f\t - %s\f", conf.loadThreshold, loadThresholdHelp)

		log.SetLevel(log.DebugLevel)
	}
}

func middleware(h http.Handler) http.Handler {
	classifier := shaper.ClassifierFunc(func(r *http.Request) shaper.Class {
		if deciders.IsCrawler(r) {
			return shaper.Class(2)
		} else {
			return shaper.Class(1)
		}
	})

	green := shaper.NewGreen(h)
	s := shaper.New(classifier, green)
	s.BucketRollingWindow = conf.bucketRollingWindow
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, err := utilisation()
		if err != nil {
			log.Warn(err)
		} else {
			transitioned := false
			if u > conf.loadThreshold {
				_, alreadyRed := s.Regime.(*shaper.Red)
				if !alreadyRed {
					// Scale Rps to 100% load average to find out with what we can cope.
					totalRps := s.GetTotalRps()
					saneRps := totalRps / u
					class1Share := s.GetClassRps(1) / totalRps
					// Cap the preferred share at 80% to not starve out class 2 completely.
					if class1Share > 0.8 {
						class1Share = 0.8
					}
					red := shaper.NewRed(h, rate.Limit(saneRps*class1Share))
					s.Regime = red
					transitioned = true
				}
			} else {
				_, alreadyGreen := s.Regime.(*shaper.Green)
				if !alreadyGreen {
					s.Regime = green
					transitioned = true
				}
			}

			if transitioned {
				log.WithFields(log.Fields{
					"Utilisation": fmt.Sprintf("%.2f", u),
					"Regime":      s.Regime,
					"TotalRps":    s.GetTotalRps(),
				}).Debug("Regime set")
			}
		}

		s.ServeHTTP(w, r)
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
