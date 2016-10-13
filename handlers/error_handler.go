package handlers

import (
	"net/http"

	sberrors "github.com/atulkc/fabric-service-broker/errors"
)

func handleDBReadError(err error, w http.ResponseWriter) {
	log.Error("Error in reading from DB", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(sberrors.ErrDBRead))
}

func handleDBSaveError(err error, w http.ResponseWriter) {
	log.Error("Error in saving to DB", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(sberrors.ErrDBSave))
}

func handleDBDeleteError(err error, w http.ResponseWriter) {
	log.Error("Error in deleting from DB", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(sberrors.ErrDBDelete))
}

func handleServiceInstanceAlreadyExists(instanceId string, w http.ResponseWriter) {
	log.Infof("Service instance:%s already exists", instanceId)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(sberrors.ErrResourceAlreadyExists))
}

func handleServiceInstanceGone(instanceId string, w http.ResponseWriter) {
	log.Infof("Service instance:%s not found in DB", instanceId)
	w.WriteHeader(http.StatusGone)
	w.Write([]byte("{}"))
}

func handleServiceBindingGone(bindingId string, w http.ResponseWriter) {
	log.Infof("Service binding:%s not found in DB", bindingId)
	w.WriteHeader(http.StatusGone)
	w.Write([]byte("{}"))
}

func handleServiceInstanceInflight(instanceId string, w http.ResponseWriter) {
	log.Infof("Service instance is still being deployed: %s", instanceId)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(sberrors.ErrProvisionInFlight))
}

func handleOutOfNetworks(w http.ResponseWriter) {
	log.Error("No networks available for deployment")
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write([]byte(sberrors.ErrNetworksUnavailable))
}

func handleManifestGenerationError(err error, w http.ResponseWriter) {
	log.Error("Error in generating manifest for deployment", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(sberrors.ErrManifestGeneration))
}

func handleInternalServerError(err error, w http.ResponseWriter) {
	log.Error("Unexpected error occurred", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

func handleBoshConnectError(err error, w http.ResponseWriter) {
	log.Error("Error connecting to Bosh", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(sberrors.ErrBoshConnect))
}

func handleBadRequest(errString string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(errString))
}

func handleNotFound(errString string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(errString))
}

func handleServiceBindingAlreadyExists(bindingId string, w http.ResponseWriter) {
	log.Infof("Service binding:%s already exists", bindingId)
	w.WriteHeader(http.StatusConflict)
	w.Write([]byte(sberrors.ErrResourceAlreadyExists))
}

func handleInstanceAlreadyBound(w http.ResponseWriter) {
	log.Infof("Binding exists for service instance")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(sberrors.ErrBindingsExist))
}

func handleProvisioningError(w http.ResponseWriter, deploymentName string, planId string) {
	log.Infof("Provisioning request made for dedicated plan :%s and pre-existing deployment :%s", 
			  planId, deploymentName)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(sberrors.ErrBindingsExist))
}

func handleDeprovisioningError(w http.ResponseWriter, deploymentName string, planId string, numInstances int) {
	log.Infof("Derovisioning request made with %d bound instances in shared plan :%s deployment :%s", 
			  numInstances, planId, deploymentName)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(sberrors.ErrBindingsExist))
}
