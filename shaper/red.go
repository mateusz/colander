package shaper

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/time/rate"
)

type Red struct {
	Handler     http.Handler
	limiter     *rate.Limiter
	rps         rate.Limit
	refuseCount int
}

func NewRed(h http.Handler, r rate.Limit) *Red {
	return &Red{
		Handler: h,
		limiter: rate.NewLimiter(r, 1),
		rps:     r,
	}
}

func (g *Red) String() string {
	return fmt.Sprintf("Red(%f)", g.rps)
}

func (red *Red) ShapeHTTP(b *Bucket, w http.ResponseWriter, r *http.Request) {
	if b.Class == 2 {
		if ok := red.limiter.Allow(); !ok {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(http.StatusText(http.StatusServiceUnavailable)))
			red.refuseCount++
			return
		}
	}
	red.Handler.ServeHTTP(w, r)

	var acceptRatio float64
	if b.Class == 1 {
		acceptRatio = 1.0
	} else {
		if red.refuseCount == 0 {
			acceptRatio = 0.0
		} else {
			acceptRatio = float64(1) / float64(red.refuseCount)
		}
	}
	log.WithFields(log.Fields{
		"Class":       b.Class,
		"AcceptRatio": fmt.Sprintf("%.2f", acceptRatio),
		"Regime":      red,
	}).Debug("Request processed")
	red.refuseCount = 0
}
