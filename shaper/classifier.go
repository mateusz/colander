package shaper

import "net/http"

type Class int

type Classifier interface {
	GetClass(r *http.Request) Class
}

type ClassifierFunc func(r *http.Request) Class

func (f ClassifierFunc) GetClass(r *http.Request) Class {
	return f(r)
}
