package deciders

import "net/http"

type Decider interface {
	Belongs(r *http.Request) bool
}
