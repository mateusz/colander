package shaper

import "net/http"

type Regime interface {
	ShapeHTTP(b *Bucket, w http.ResponseWriter, r *http.Request)
}
