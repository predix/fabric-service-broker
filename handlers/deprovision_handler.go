package handlers

import (
	"net/http"

	"github.com/atulkc/fabric-service-broker/errors"
)

var deprovisionResponse = `
{
 "operation": "task10"
}
`

func DeprovisioningHandler(w http.ResponseWriter, r *http.Request) {
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
