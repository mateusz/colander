package regimes

import "net/http"

// TODO It would possibly be a good idea to remove Bucket from here
// so this could become http.Handler. A set of buckets could be
// configured upon Regime construction ... umm
// but for that to work we'd probably need to move classification
// into the regime, because it would need to be able to figure out
// the traffic class per request internally.
// Maybe a classifier.Decider object could be passed by the programmer
// consuming the Regime, and the decider would return the traffic class?
// Or maybe the decider would even return a bucket?
type Regime interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request, b *Bucket)
}
