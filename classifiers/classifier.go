package classifiers

import "net/http"

type Classifier interface {
	Belongs(r *http.Request) bool
}
