package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/atulkc/fabric-service-broker/db"
	dbmodels "github.com/atulkc/fabric-service-broker/db/models"
	sberrors "github.com/atulkc/fabric-service-broker/errors"
	"github.com/atulkc/fabric-service-broker/models"
	"github.com/atulkc/fabric-service-broker/util"
	"github.com/gorilla/mux"
)

type ServiceLifecycleHandler interface {
	Provision(w http.ResponseWriter, r *http.Request)
	Deprovision(w http.ResponseWriter, r *http.Request)
	LastOperation(w http.ResponseWriter, r *http.Request)
	Bind(w http.ResponseWriter, r *http.Request)
	Unbind(w http.ResponseWriter, r *http.Request)
}

var asyncResponse = `
{
 "operation": "%d"
}
`

type slHandler struct {
	boshDetails       *models.BoshDetails
	modelsRepo        db.ModelsRepo
	availableNetworks map[string]struct{}
	boshClient        util.BoshClient
	lock              *sync.Mutex
}

func NewServiceLifecycleHandler(repo db.ModelsRepo, boshClient util.BoshClient, boshDetails *models.BoshDetails) ServiceLifecycleHandler {

	s := &slHandler{
		boshDetails:       boshDetails,
		modelsRepo:        repo,
		boshClient:        boshClient,
		availableNetworks: make(map[string]struct{}),
		lock:              &sync.Mutex{},
	}

	s.RefreshAvailableNetworks()

	return s
}

func (s *slHandler) RefreshAvailableNetworks() {
	log.Debug("Refreshing available networks")

	s.lock.Lock()
	defer func() {
		s.lock.Unlock()
		log.Debugf("Available networks: %s", s.availableNetworks)
	}()

	for _, networkName := range s.boshDetails.NetworkNames {
		s.availableNetworks[networkName] = struct{}{}
	}

	serviceInstances, err := s.modelsRepo.ListServiceInstances()
	if err != nil {
		log.Error("Unable to fetch service instances from db", err)
		return
	}

	for _, serviceInstance := range serviceInstances {
		delete(s.availableNetworks, serviceInstance.NetworkName)
	}
}

func (s *slHandler) Provision(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Info("Handling PUT /v2/service_instances")
	vars := mux.Vars(r)
	instanceId := vars["instanceId"]

	if !s.isAsyncRequest(w, r) {
		return
	}

	existingServiceInstance, err := s.modelsRepo.FindServiceInstance(instanceId)
	if err != nil {
		handleDBReadError(err, w)
		return
	}
	if existingServiceInstance != nil {
		handleServiceInstanceAlreadyExists(instanceId, w)
		return
	}

	// Get the first available network
	var networkName string
	for netName, _ := range s.availableNetworks {
		networkName = netName
	}

	if networkName == "" {
		handleOutOfNetworks(w)
		return
	}
	log.Debugf("Network name selected for this deployment: %s", networkName)

	deploymentName := fmt.Sprintf("fabric-%s", instanceId)

	manifest, err := models.NewManifest(deploymentName, networkName, s.boshDetails)
	if err != nil {
		handleManifestGenerationError(err, w)
		return
	}
	log.Debugf("Manifest generated for deployment")

	task, err := s.boshClient.CreateDeployment(*manifest)
	if err != nil {
		handleInternalServerError(err, w)
		return
	}

	serviceInstance := dbmodels.ServiceInstance{
		BaseModel:           dbmodels.BaseModel{Id: instanceId},
		DeploymentName:      deploymentName,
		NetworkName:         networkName,
		BlockchainNetworkId: instanceId,
		ProvisionTaskId:     strconv.Itoa(task.Id),
		DeprovisionTaskId:   "",
	}
	err = s.modelsRepo.UpsertServiceInstance(serviceInstance)
	if err != nil {
		handleDBSaveError(err, w)
		return
	}
	log.Debugf("Service instance saved to DB")
	delete(s.availableNetworks, networkName)
	log.Debugf("Network %s deleted from available networks", networkName)

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(fmt.Sprintf(asyncResponse, task.Id)))
}

func (s *slHandler) Deprovision(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Info("Handling DELETE /v2/service_instances")

	if !s.isAsyncRequest(w, r) {
		return
	}

	vars := mux.Vars(r)
	instanceId := vars["instanceId"]
	serviceInstance, err := s.modelsRepo.FindServiceInstance(instanceId)
	if err != nil {
		handleDBReadError(err, w)
		return
	}
	if serviceInstance == nil {
		handleServiceInstanceGone(instanceId, w)
		return
	}

	isProvisionComplete, err := s.isProvisionComplete(serviceInstance)
	if err != nil {
		handleBoshConnectError(err, w)
		return
	}
	if !isProvisionComplete {
		handleServiceInstanceInflight(instanceId, w)
		return
	}

	bindings, err := s.modelsRepo.AssociatedServiceBindings(instanceId)
	if err != nil {
		handleDBReadError(err, w)
		return
	}
	if len(bindings) > 0 {
		log.Infof("Total %d bindings exist for service instance %s", len(bindings), instanceId)
		handleInstanceAlreadyBound(w)
		return
	}
	log.Debugf("No bindings for service instance :%d", instanceId)

	task, err := s.boshClient.DeleteDeployment(serviceInstance.DeploymentName)
	if err != nil {
		handleInternalServerError(err, w)
		return
	}

	serviceInstance.DeprovisionTaskId = strconv.Itoa(task.Id)

	err = s.modelsRepo.UpsertServiceInstance(*serviceInstance)
	if err != nil {
		handleDBSaveError(err, w)
		return
	}
	log.Debug("Saved service instance to DB")

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(fmt.Sprintf(asyncResponse, task.Id)))
}

func (s *slHandler) LastOperation(w http.ResponseWriter, r *http.Request) {
	log.Info("Handling GET /v2/service_instances/:instanceId/last_operation")
	query := r.URL.Query()
	taskId := query["operation"]
	if len(taskId) < 1 {
		handleBadRequest("No operation parameter specified", w)
		return
	}

	vars := mux.Vars(r)
	instanceId := vars["instanceId"]
	serviceInstance, err := s.modelsRepo.FindServiceInstance(instanceId)
	if err != nil {
		handleDBReadError(err, w)
		return
	}
	if serviceInstance == nil {
		handleServiceInstanceGone(instanceId, w)
		return
	}

	log.Debugf("Checking status of task:%s", taskId[0])

	task, err := s.boshClient.GetTask(taskId[0])
	if err != nil {
		handleInternalServerError(err, w)
		return
	}

	operation := models.OpProvision
	if serviceInstance.DeprovisionTaskId == taskId[0] {
		operation = models.OpDeprovision
	}

	lastOperationResponse := models.GetLastOperationResponse(operation, task.State)
	if lastOperationResponse.State == models.StateSucceeded &&
		taskId[0] == serviceInstance.DeprovisionTaskId {
		log.Info("Delete operation succeeded. Removing entry from DB")
		s.lock.Lock()
		defer s.lock.Unlock()
		_, err := s.modelsRepo.DeleteServiceInstance(serviceInstance.Id)
		if err != nil {
			handleDBDeleteError(err, w)
			return
		}
		log.Infof("Returning network %s back to available pool", serviceInstance.NetworkName)
		s.availableNetworks[serviceInstance.NetworkName] = struct{}{}
	}
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(lastOperationResponse)
}

func (s *slHandler) Bind(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Info("Handling PUT /v2/service_instances/:instanceId/service_bindings/:bindingId")

	vars := mux.Vars(r)
	instanceId := vars["instanceId"]
	bindingId := vars["bindingId"]

	serviceInstance, err := s.modelsRepo.FindServiceInstance(instanceId)
	if err != nil {
		handleDBReadError(err, w)
		return
	}
	if serviceInstance == nil {
		handleNotFound("instances not found", w)
		return
	}

	serviceBinding, err := s.modelsRepo.FindServiceBinding(bindingId)
	if err != nil {
		handleDBReadError(err, w)
		return
	}

	if serviceBinding != nil {
		handleServiceBindingAlreadyExists(bindingId, w)
		return
	}
	log.Debugf("Deployment name for instance:%s is %s", instanceId, serviceInstance.DeploymentName)

	isProvisionComplete, err := s.isProvisionComplete(serviceInstance)
	if err != nil {
		handleBoshConnectError(err, w)
		return
	}

	if !isProvisionComplete {
		handleServiceInstanceInflight(instanceId, w)
		return
	}

	vmsIps, err := s.boshClient.GetVmIps(serviceInstance.DeploymentName)
	if err != nil {
		log.Error("Error in getting VM details", err)
		handleInternalServerError(err, w)
		return
	}

	serviceBinding = &dbmodels.ServiceBinding{
		BaseModel:         dbmodels.BaseModel{Id: bindingId},
		ServiceInstanceId: instanceId,
	}

	err = s.modelsRepo.UpsertServiceBinding(*serviceBinding)
	if err != nil {
		handleDBSaveError(err, w)
		return
	}

	s.writeBindingResponse(vmsIps, w)
}

func (s *slHandler) Unbind(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Info("Handling DELETE /v2/service_instances/:instanceId/service_bindings/:bindingId")

	vars := mux.Vars(r)
	instanceId := vars["instanceId"]
	bindingId := vars["bindingId"]

	serviceInstance, err := s.modelsRepo.FindServiceInstance(instanceId)
	if err != nil {
		handleDBReadError(err, w)
		return
	}
	if serviceInstance == nil {
		handleNotFound("instance not found", w)
		return
	}

	serviceBinding, err := s.modelsRepo.FindServiceBinding(bindingId)
	if err != nil {
		handleDBReadError(err, w)
		return
	}

	if serviceBinding == nil {
		handleServiceBindingGone(bindingId, w)
		return
	}

	log.Debugf("Deleting binding :%s from DB", bindingId)
	_, err = s.modelsRepo.DeleteServiceBinding(bindingId)
	if err != nil {
		handleDBSaveError(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func (s *slHandler) writeBindingResponse(vmsIps map[string][]string, w http.ResponseWriter) {
	peerIps := vmsIps["peer"]

	peerEndpoints := make([]string, 0)

	for _, peerIp := range peerIps {
		peerEndpoints = append(peerEndpoints, fmt.Sprintf("%s:5000", peerIp))
	}

	bindCredentials := models.BindCredentials{
		Credentials: models.BlockChainCredentials{
			PeerEndpoints: peerEndpoints,
		},
	}
	log.Debugf("Created binding crendentials:%#v", bindCredentials)

	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	encoder.Encode(bindCredentials)
}

func (s *slHandler) isAsyncRequest(w http.ResponseWriter, r *http.Request) bool {
	query := r.URL.Query()
	async := query["accepts_incomplete"]
	if len(async) < 1 || async[0] != "true" {
		w.WriteHeader(422)
		w.Write([]byte(sberrors.ErrAsyncResponse))
		return false
	}

	return true
}

func (s *slHandler) isProvisionComplete(serviceInstance *dbmodels.ServiceInstance) (bool, error) {
	task, err := s.boshClient.GetTask(serviceInstance.ProvisionTaskId)
	if err != nil {
		return false, err
	}
	if task.State != models.BoshStateDone {
		return false, nil
	}
	return true, nil
}
