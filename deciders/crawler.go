package decider

import (
	"log"
	"net/http"
)

type Crawler struct {
	logger *log.Logger
}

func NewCrawler(l *log.Logger) *Crawler {
	c := &Crawler{
		logger: l,
	}
	return c
}

func (c *Crawler) Belongs(r *http.Request) bool {
	uaList, ok := r.Header["User-Agent"]
	if !ok {
		return false
	}
	if len(uaList) == 0 {
		return false
	}
	ua := uaList[0]
	if ua == "crawler" {
		return true
	}
	return false
}
