package handlers

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/atulkc/fabric-service-broker/db"
	sberrors "github.com/atulkc/fabric-service-broker/errors"
	"github.com/atulkc/fabric-service-broker/schema"
	"github.com/gorilla/mux"
)

type ServiceLifecycleHandler interface {
	Provision(w http.ResponseWriter, r *http.Request)
	Deprovision(w http.ResponseWriter, r *http.Request)
	LastOperation(w http.ResponseWriter, r *http.Request)
}

var provisionResponse = `
{
 "operation": "%s"
}
`
var deprovisionResponse = `
{
 "operation": "task10"
}
`

type slHandler struct {
	boshDetails         *schema.BoshDetails
	serviceInstanceRepo db.ServiceInstanceRepo
	availableNetworks   map[string]struct{}
	httpClient          *http.Client
	lock                *sync.Mutex
}

func NewHttpClient(skipTLSVerification bool) *http.Client {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: skipTLSVerification,
	}
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errors.New("No redirects")
		},
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).Dial,
			TLSClientConfig: tlsConfig,
		},
	}
}

func NewServiceLifecycleHandler(boshDetails *schema.BoshDetails) ServiceLifecycleHandler {
	inMem := db.GetInMemoryDB()

	s := &slHandler{
		boshDetails:         boshDetails,
		serviceInstanceRepo: inMem,
		httpClient:          NewHttpClient(true),
		availableNetworks:   make(map[string]struct{}),
		lock:                &sync.Mutex{},
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

	serviceInstances, err := s.serviceInstanceRepo.List()
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

	query := r.URL.Query()
	async := query["accepts_incomplete"]
	if len(async) < 1 || async[0] != "true" {
		w.WriteHeader(422)
		w.Write([]byte(sberrors.ErrAsyncResponse))
		return
	}

	// Get the first available network
	var networkName string
	for netName, _ := range s.availableNetworks {
		networkName = netName
	}

	if networkName == "" {
		log.Error("No networks available for deployment")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(sberrors.ErrNetworksUnavailable))
		return
	}
	log.Debugf("Network name selected for this deployment: %s", networkName)

	deploymentName := fmt.Sprintf("fabric-%s", instanceId)

	manifest, err := schema.NewManifest(deploymentName, networkName, s.boshDetails)
	if err != nil {
		log.Error("Error in generating manifest for deployment", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(sberrors.ErrManifestGeneration))
		return
	}
	log.Debugf("Manifest generated for deployment")

	body := manifest.String()
	log.Debugf("Manifest for deployment:%s", body)
	url := fmt.Sprintf("%s%s", s.boshDetails.BoshDirectorUrl, "/deployments")
	request, err := http.NewRequest("POST", url, bytes.NewReader([]byte(body)))
	if err != nil {
		log.Error("Error in creating http request", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(sberrors.ErrHttpRequest))
		return
	}
	request.Header.Set("Content-Type", "text/yaml")
	log.Debugf("Http request for BOSH director created")

	resp, err := s.httpClient.Do(request)
	if err != nil && !strings.Contains(err.Error(), "No redirects") {
		log.Error("Error in connecting to Bosh", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(sberrors.ErrBoshConnect))
		return
	}

	taskUrl := resp.Header.Get("Location")
	if taskUrl == "" {
		log.Error("Invalid response from Bosh", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(sberrors.ErrBoshInvalidResponse))
		return
	}

	split := strings.Split(taskUrl, "/")
	taskId := split[len(split)-1]
	if taskId == "" {
		log.Error("Invalid response from Bosh", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(sberrors.ErrBoshInvalidResponse))
		return
	}

	log.Infof("Successfully initiated deployment:%s. Task Id is: %s", deploymentName, taskId)

	serviceInstance := db.ServiceInstance{
		InstanceId:          instanceId,
		DeploymentName:      deploymentName,
		NetworkName:         networkName,
		BlockchainNetworkId: instanceId,
		ProvisionTaskId:     taskId,
		DeprovisionTaskId:   "",
	}
	err = s.serviceInstanceRepo.Create(serviceInstance)
	if err != nil {
		log.Error("Error in saving to DB", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(sberrors.ErrDBSave))
		return
	}
	log.Debugf("Service instance saved to DB")
	delete(s.availableNetworks, networkName)
	log.Debugf("Network %s deleted from available networks", networkName)

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(fmt.Sprintf(provisionResponse, taskId)))
}

func (s *slHandler) Deprovision(w http.ResponseWriter, r *http.Request) {
	log.Info("Handling DELETE /v2/service_instances")
	// vars := mux.Vars(r)
	// instanceId := vars["instanceId"]

	query := r.URL.Query()
	async := query["accepts_incomplete"]
	if len(async) < 1 || async[0] != "true" {
		w.WriteHeader(422)
		w.Write([]byte(sberrors.ErrAsyncResponse))
		return
	}

	// TODO: get service_id and plan_id
	// 410 if it doesn't exist

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(deprovisionResponse))
}

func (s *slHandler) LastOperation(w http.ResponseWriter, r *http.Request) {
	log.Info("Handling GET /v2/service_instances/:instanceId/last_operation")
	query := r.URL.Query()
	taskId := query["operation"]
	if len(taskId) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No operation parameter specified"))
		return
	}

	vars := mux.Vars(r)
	instanceId := vars["instanceId"]
	serviceInstance, err := s.serviceInstanceRepo.Find(instanceId)
	if err != nil {
		log.Error("Error in reading from DB", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(sberrors.ErrDBRead))
		return
	}
	if serviceInstance == nil {
		log.Infof("Service instance:%s not found in DB", instanceId)
		w.WriteHeader(http.StatusGone)
		w.Write([]byte("{}"))
		return
	}

	log.Debugf("Checking status of task:%s", taskId[0])
	url := fmt.Sprintf("%s%s%s", s.boshDetails.BoshDirectorUrl, "/tasks/", taskId[0])
	resp, err := s.httpClient.Get(url)
	if err != nil && !strings.Contains(err.Error(), "No redirects") {
		log.Error("Error in connecting to Bosh", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(sberrors.ErrBoshConnect))
		return
	}
	log.Debug("Received response from Bosh")
	if resp.StatusCode == http.StatusNotFound {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid operation parameter specified"))
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Error("Non OK status code from BOSH")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(sberrors.ErrBoshInvalidResponse))
		return
	}

	task := schema.Task{}
	err = json.NewDecoder(resp.Body).Decode(&task)
	if err != nil {
		log.Error("Error in decoding response from Bosh", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(sberrors.ErrBoshInvalidResponse))
		return
	}

	lastOperationResponse := schema.GetLastOperationResponse(task.State)
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(lastOperationResponse)
}
