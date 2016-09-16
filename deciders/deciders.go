package deciders

import "net/http"

func IsCrawler(r *http.Request) bool {
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
