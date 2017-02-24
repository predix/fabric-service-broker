package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/predix/fabric-service-broker/bosh"
	"github.com/predix/fabric-service-broker/db"
	"github.com/predix/fabric-service-broker/db/models"
	sberrors "github.com/predix/fabric-service-broker/errors"
	"github.com/predix/fabric-service-broker/rest_models"
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
	boshDetails       *bosh.Details
	modelsRepo        db.ModelsRepo
	availableNetworks map[string]struct{}
	boshClient        bosh.Client
	lock              *sync.Mutex
}

func NewServiceLifecycleHandler(repo db.ModelsRepo, boshClient bosh.Client, boshDetails *bosh.Details) ServiceLifecycleHandler {

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

// We are locking the mutex for entire operation essentially serializing
// access to any rest endpoint on this service broker. This can be improved
// to lock only network name selection using some form of DB locking but for
// MVP this serialized version should be fine.
// Another shortcoming of this approach is that this locking is not cluster safe
// meaning if there are multiple instances of this server running then they dont
// coordinate with each other in selection of network name. This is also something
// that we need to improve on past MVP.
func (s *slHandler) Provision(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Info("Handling PUT /v2/service_instances")
	vars := mux.Vars(r)
	instanceId := vars["instanceId"]

	if !s.isAsyncRequest(w, r) {
		return
	}

	decoder := json.NewDecoder(r.Body)

	var serviceProvisionRequest rest_models.ServiceProvisionRequest
	err := decoder.Decode(&serviceProvisionRequest)
	if err != nil {
		handleBadRequest(err.Error(), w)
		return
	}

	if !s.isValidServiceIdAndPlanId(serviceProvisionRequest.ServiceId, serviceProvisionRequest.PlanId, w) {
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
	permissioned := s.isPermissioned(serviceProvisionRequest.PlanId)

	manifest, err := bosh.NewManifest(deploymentName, networkName, permissioned, s.boshDetails)
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

	serviceInstance := models.ServiceInstance{
		BaseModel:           models.BaseModel{Id: instanceId},
		ServiceId:           serviceProvisionRequest.ServiceId,
		PlanId:              serviceProvisionRequest.PlanId,
		OrganizationGuid:    serviceProvisionRequest.OrganizationGuid,
		SpaceGuid:           serviceProvisionRequest.SpaceGuid,
		DeploymentName:      deploymentName,
		NetworkName:         networkName,
		BlockchainNetworkId: instanceId,
		ProvisionTaskId:     strconv.Itoa(task.Id),
		DeprovisionTaskId:   "",
	}
	err = s.modelsRepo.CreateServiceInstance(serviceInstance)
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

	err = s.modelsRepo.UpdateServiceInstance(*serviceInstance)
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

	operation := rest_models.OpProvision
	if serviceInstance.DeprovisionTaskId == taskId[0] {
		operation = rest_models.OpDeprovision
	}

	lastOperationResponse := rest_models.GetLastOperationResponse(operation, task.State)
	if lastOperationResponse.State == rest_models.StateSucceeded &&
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

	decoder := json.NewDecoder(r.Body)

	var serviceBindingRequest rest_models.ServiceBindingRequest
	err := decoder.Decode(&serviceBindingRequest)
	if err != nil {
		handleBadRequest(err.Error(), w)
		return
	}

	if !s.isValidServiceIdAndPlanId(serviceBindingRequest.ServiceId, serviceBindingRequest.PlanId, w) {
		return
	}

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

	serviceBinding = &models.ServiceBinding{
		BaseModel:         models.BaseModel{Id: bindingId},
		ServiceInstanceId: instanceId,
		AppId:             serviceBindingRequest.AppGuid,
	}

	err = s.modelsRepo.CreateServiceBinding(*serviceBinding)
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

	bindCredentials := rest_models.BindCredentials{
		Credentials: rest_models.BlockChainCredentials{
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

func (s *slHandler) isProvisionComplete(serviceInstance *models.ServiceInstance) (bool, error) {
	task, err := s.boshClient.GetTask(serviceInstance.ProvisionTaskId)
	if err != nil {
		return false, err
	}
	if task.State != bosh.BoshStateDone {
		return false, nil
	}
	return true, nil
}

func (s *slHandler) isValidServiceIdAndPlanId(serviceId, planId string, w http.ResponseWriter) bool {
	if serviceId != rest_models.DefaultServiceId {
		log.Errorf("Invalid service id:%s specified", serviceId)
		handleBadRequest("Invalid Service Id", w)
		return false
	}
	if planId != rest_models.PermissionlessPlanId && planId != rest_models.PermissionedPlanId {
		log.Errorf("Invalid plan id:%s specified", planId)
		handleBadRequest("Invalid Plan Id", w)
		return false
	}
	return true
}

func (s *slHandler) isPermissioned(planId string) bool {
	if planId == rest_models.PermissionedPlanId {
		return true
	}
	return false
}
