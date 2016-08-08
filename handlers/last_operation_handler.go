package handlers

import "net/http"

var lastResponse = `
{
	"state": "succeeded",
	"description": "created fabric cluster"
}
`

func LastOperationHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Handling GET /v2/service_instances/:instanceId/last_operation")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(lastResponse))
}
