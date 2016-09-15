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

func (s *Shaper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	class := s.Classifier.GetClass(r)

	if _, ok := s.buckets[class]; !ok {
		s.bucketSync.Lock()
		s.buckets[class] = NewBucket(class, s.BucketRollingWindow)
		s.bucketSync.Unlock()

		log.WithFields(log.Fields{
			"Class":               class,
			"BucketRollingWindow": s.BucketRollingWindow,
		}).Debug("Class created")
	}
	s.Regime.ShapeHTTP(s.buckets[class], w, r)
}
