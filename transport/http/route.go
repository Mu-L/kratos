package http

import (
	"net/http"
	"path"

	"github.com/gorilla/mux"
)

// RouteGroup .
type RouteGroup struct {
	root   string
	router *mux.Router
}

// GET .
func (r *RouteGroup) GET(p string, h http.HandlerFunc) {
	r.router.HandleFunc(path.Join(r.root, p), h).Methods("GET")
}

// HEAD .
func (r *RouteGroup) HEAD(p string, h http.HandlerFunc) {
	r.router.HandleFunc(path.Join(r.root, p), h).Methods("HEAD")
}

// POST .
func (r *RouteGroup) POST(p string, h http.HandlerFunc) {
	r.router.HandleFunc(path.Join(r.root, p), h).Methods("POST")
}

// PUT .
func (r *RouteGroup) PUT(p string, h http.HandlerFunc) {
	r.router.HandleFunc(path.Join(r.root, p), h).Methods("PUT")
}

// DELETE .
func (r *RouteGroup) DELETE(p string, h http.HandlerFunc) {
	r.router.HandleFunc(path.Join(r.root, p), h).Methods("DELETE")
}

// PATCH .
func (r *RouteGroup) PATCH(p string, h http.HandlerFunc) {
	r.router.HandleFunc(path.Join(r.root, p), h).Methods("PATCH")
}

// OPTIONS .
func (r *RouteGroup) OPTIONS(p string, h http.HandlerFunc) {
	r.router.HandleFunc(path.Join(r.root, p), h).Methods("OPTIONS")
}
