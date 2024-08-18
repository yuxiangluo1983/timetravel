package api

import (
	"github.com/gorilla/mux"
	"github.com/yuxiangluo1983/timetravel/service"
)

type API struct {
	records service.RecordService
}

func NewAPI(records service.RecordService) *API {
	return &API{records}
}

// generates all api routes
func (a *API) CreateRoutes(routes *mux.Router) {
	routes.Path("/records/{id}").HandlerFunc(a.GetRecords).Methods("GET")
	routes.Path("/records/{id}").HandlerFunc(a.PostRecords).Methods("POST")
}

// generates all api v2 routes
func (a *API) CreateV2Routes(routes *mux.Router) {
	routes.Path("/records/{id}").Queries("versions", "{versions}").HandlerFunc(a.GetRecordsV2).Methods("GET")
	routes.Path("/records/{id}").HandlerFunc(a.PostRecordsV2).Methods("POST")
}
