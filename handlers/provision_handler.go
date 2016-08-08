package handlers

import (
	"net/http"

	"github.com/atulkc/fabric-service-broker/errors"
)

var provisionResponse = `
{
 "dashboardurl": "http://example-dashboard.example.com/9189kdfsk0vfnku",
 "operation": "task10"
}
`

func ProvisioningHandler(w http.ResponseWriter, r *http.Request) {
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

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(provisionResponse))
}
