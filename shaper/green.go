package shaper

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Green is counting-only regime which should be used for
// establishing the baseline RpsAvg and RespTimeAvg.
// It is not intended for any traffic shaping, and should only
// be used if the server is under "normal" operating conditions,
// i.e. is not under load.
type Green struct {
	Handler             http.Handler
	Verbose             bool
	countWindowStart    time.Time
	reqCount            int64
	countWindowDuration time.Duration
}

func NewGreen(h http.Handler) *Green {
	return &Green{
		Handler:             h,
		Verbose:             false,
		countWindowStart:    time.Now(),
		reqCount:            0,
		countWindowDuration: time.Second,
	}
}

func (g *Green) ShapeHTTP(b *Bucket, w http.ResponseWriter, r *http.Request) {
	// One Rps sample is obtained by calculating amount of requests in the window,
	// where the window is at minimum g.countWindowDuration. This could lead to
	// odd sampling if requests are infrequent, in which case the average will take longer
	// to compute (but it should still be correct).
	d := time.Since(g.countWindowStart)
	if d > g.countWindowDuration {
		dSeconds := float64(d) / float64(time.Second)
		rps := float64(g.reqCount+1) / dSeconds
		b.addRpsSample(rps)
		// TODO this really should use a mutex
		g.countWindowStart = time.Now()
		g.reqCount = 0
	} else {
		g.reqCount++
	}

	start := time.Now()
	g.Handler.ServeHTTP(w, r)
	reqTime := time.Since(start)
	b.addRespTimeSample(reqTime)

	log.WithFields(log.Fields{
		"Class":       b.Class,
		"RpsAvg":      b.RpsAvg,
		"RespTimeAvg": b.RespTimeAvg,
	}).Debug("Request processed")
}
