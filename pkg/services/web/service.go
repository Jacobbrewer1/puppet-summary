package web

import (
	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/gorilla/mux"
	"net/http"
)

type service struct {
	r  *mux.Router
	db dataaccess.Database
}

func NewService(db dataaccess.Database) http.Handler {
	r := mux.NewRouter()
	return NewServiceFromRouter(r, db, nil)
}

func NewServiceFromRouter(r *mux.Router, db dataaccess.Database, middlewareFunc func(handler http.HandlerFunc) http.HandlerFunc) http.Handler {
	svc := &service{
		r:  r,
		db: db,
	}

	if middlewareFunc == nil {
		middlewareFunc = func(handler http.HandlerFunc) http.HandlerFunc {
			return handler
		}
	}

	r.HandleFunc(pathIndex, middlewareFunc(svc.indexHandler)).Methods(http.MethodGet)
	r.HandleFunc(pathIndexEnv, middlewareFunc(svc.indexHandler)).Methods(http.MethodGet)
	r.HandleFunc(pathNodeFqdn, middlewareFunc(svc.nodeFqdnHandler)).Methods(http.MethodGet)
	r.HandleFunc(pathReportID, middlewareFunc(svc.reportIDHandler)).Methods(http.MethodGet)

	return r
}
