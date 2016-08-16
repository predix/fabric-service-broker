package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/atulkc/fabric-service-broker/rest_models"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("handler")

func CatalogHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Serving /v2/catalog")
	serviceCatalog := rest_models.ServiceCatalog{
		Services: rest_models.Services{rest_models.GetDefaultService()},
	}
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(serviceCatalog)
}
