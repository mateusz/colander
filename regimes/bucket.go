package regimes

import (
	"time"

	"golang.org/x/time/rate"
)

// Bucket keeps track of all state per a single traffic class.
// This information is required for the Regime to operate.
type Bucket struct {
	Class         int
	RollingWindow int
	RpsAvg        float64
	RespTimeAvg   time.Duration
	Limiter       *rate.Limiter
}

func NewBucket(c int, rw int) *Bucket {
	return &Bucket{
		Class:         c,
		RollingWindow: rw,
		RpsAvg:        0.0,
		RespTimeAvg:   time.Second,
	}
}

func (b *Bucket) addRpsSample(s float64) {
	b.RpsAvg -= b.RpsAvg / float64(b.RollingWindow)
	b.RpsAvg += s / float64(b.RollingWindow)
}

func (b *Bucket) addRespTimeSample(s time.Duration) {
	b.RespTimeAvg -= b.RespTimeAvg / time.Duration(b.RollingWindow)
	b.RespTimeAvg += s / time.Duration(b.RollingWindow)
}
