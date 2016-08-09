package handlers

import (
	"net/http"

	"github.com/atulkc/fabric-service-broker/errors"
)

type ServiceLifecycleHandler interface {
	Provision(w http.ResponseWriter, r *http.Request)
	Deprovision(w http.ResponseWriter, r *http.Request)
}

var provisionResponse = `
{
 "dashboardurl": "http://example-dashboard.example.com/9189kdfsk0vfnku",
 "operation": "task10"
}
`
var deprovisionResponse = `
{
 "operation": "task10"
}
`

type slHandler struct {
}

func NewServiceLlifecycleHandler() ServiceLifecycleHandler {
	return &slHandler{}
}

func (s *slHandler) Provision(w http.ResponseWriter, r *http.Request) {
	log.Info("Handling PUT /v2/service_instances")
	// vars := mux.Vars(r)
	// instanceId := vars["instanceId"]

	query := r.URL.Query()
	async := query["accepts_incomplete"]
	if len(async) < 1 || async[0] != "true" {
		w.WriteHeader(422)
		w.Write([]byte(errors.ErrAsyncResponse))
		return
	}

	// load manifest
	//
	// change name
	// fire http call to bosh director

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(provisionResponse))
}

func (s *slHandler) Deprovision(w http.ResponseWriter, r *http.Request) {
	log.Info("Handling DELETE /v2/service_instances")
	// vars := mux.Vars(r)
	// instanceId := vars["instanceId"]

	query := r.URL.Query()
	async := query["accepts_incomplete"]
	if len(async) < 1 || async[0] != "true" {
		w.WriteHeader(422)
		w.Write([]byte(errors.ErrAsyncResponse))
		return
	}

	// TODO: get service_id and plan_id
	// 410 if it doesn't exist

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(deprovisionResponse))
}
