package db

import (
	"errors"

	"github.com/op/go-logging"
)

type ServiceInstance struct {
	InstanceId          string
	DeploymentName      string
	NetworkName         string
	BlockchainNetworkId string
	ProvisionTaskId     string
	DeprovisionTaskId   string
}

func (s ServiceInstance) Validate() error {
	if s.InstanceId == "" {
		return errors.New("InstanceId cannot be empty")
	}
	if s.DeploymentName == "" {
		return errors.New("DeploymentName cannot be empty")
	}
	if s.NetworkName == "" {
		return errors.New("NetworkName cannot be empty")
	}
	if s.BlockchainNetworkId == "" {
		return errors.New("BlockchainNetworkId cannot be empty")
	}
	return nil
}

type ServiceInstanceRepo interface {
	Upsert(serviceInstance ServiceInstance) error
	Find(serviceInstanceId string) (*ServiceInstance, error)
	List() ([]ServiceInstance, error)
	Delete(serviceInstanceId string) (*ServiceInstance, error)
}

type inMemoryDb struct {
	data map[string]ServiceInstance
}

var log = logging.MustGetLogger("db")

var inMemoryDbInstance *inMemoryDb

func init() {
	inMemoryDbInstance = &inMemoryDb{
		data: make(map[string]ServiceInstance),
	}
}

func GetInMemoryDB() *inMemoryDb {
	return inMemoryDbInstance
}

func (d *inMemoryDb) Upsert(serviceInstance ServiceInstance) error {
	log.Infof("UpsertServiceInstance: %s", serviceInstance.InstanceId)
	log.Debugf("Body: %#v", serviceInstance)

	err := serviceInstance.Validate()
	if err != nil {
		return err
	}

	d.data[serviceInstance.InstanceId] = serviceInstance
	return nil
}

func (d *inMemoryDb) Find(serviceInstanceId string) (*ServiceInstance, error) {
	log.Infof("FindServiceInstance: %s", serviceInstanceId)
	serviceInstance, ok := d.data[serviceInstanceId]
	if !ok {
		log.Debugf("No record with key %s found", serviceInstanceId)
		return nil, nil
	}

	return &serviceInstance, nil
}

func (d *inMemoryDb) List() ([]ServiceInstance, error) {
	log.Infof("List")

	list := make([]ServiceInstance, len(d.data))
	for _, serviceInstance := range d.data {
		list = append(list, serviceInstance)
	}

	return list, nil
}

func (d *inMemoryDb) Delete(serviceInstanceId string) (*ServiceInstance, error) {
	log.Infof("DeleteServiceInstance: %s", serviceInstanceId)
	serviceInstance, ok := d.data[serviceInstanceId]
	if !ok {
		log.Debugf("No record with key %s found", serviceInstanceId)
	}

	delete(d.data, serviceInstanceId)

	return &serviceInstance, nil
}
