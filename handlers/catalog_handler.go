package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/op/go-logging"
	"github.com/predix/fabric-service-broker/rest_models"
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
