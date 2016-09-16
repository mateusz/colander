package shaper

import (
	"net/http"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type Shaper struct {
	Classifier          Classifier
	Regime              Regime
	BucketRollingWindow int
	buckets             map[Class]*Bucket
	bucketSync          *sync.RWMutex
}

func New(classifier Classifier, initialRegime Regime) *Shaper {
	return &Shaper{
		Classifier:          classifier,
		Regime:              initialRegime,
		BucketRollingWindow: 100,
		buckets:             make(map[Class]*Bucket),
		bucketSync:          &sync.RWMutex{},
	}
}

func (s *Shaper) GetTotalRps() float64 {
	totalRps := 0.0
	for _, b := range s.buckets {
		totalRps += b.RpsAvg
	}
	return totalRps
}

func (s *Shaper) GetClassRps(c Class) float64 {
	return s.buckets[c].RpsAvg
}

func (s *Shaper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	class := s.Classifier.GetClass(r)

	s.bucketSync.Lock()
	if _, ok := s.buckets[class]; !ok {
		s.buckets[class] = NewBucket(class, s.BucketRollingWindow)

		log.WithFields(log.Fields{
			"Class":               class,
			"BucketRollingWindow": s.BucketRollingWindow,
		}).Debug("Class created")
	}
	s.bucketSync.Unlock()

	s.Regime.ShapeHTTP(s.buckets[class], w, r)
}
