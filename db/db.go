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
	ProvisionTaskId     uint
	DeprovisionTaskId   uint
}

func (s *ServiceInstance) Validate() error {
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

type Db interface {
	CreateServiceInstance(serviceInstance ServiceInstance) error
	FindServiceInstance(serviceInstanceId string) (ServiceInstance, error)
	DeleteServiceInstance(serviceInstanceId string) (ServiceInstance, error)
}

type inMemoryDb struct {
	data map[string]*ServiceInstance
}

var log = logging.MustGetLogger("db")

var inMemoryDbInstance *inMemoryDb

func init() {
	inMemoryDbInstance = &inMemoryDb{
		data: make(map[string]*ServiceInstance),
	}
}

func GetInMemoryDB() *inMemoryDb {
	return inMemoryDbInstance
}

func (d *inMemoryDb) CreateServiceInstance(serviceInstance ServiceInstance) error {
	log.Infof("CreateServiceInstance: %s", serviceInstance.InstanceId)
	log.Debugf("Body: %#v", serviceInstance)

	err := serviceInstance.Validate()
	if err != nil {
		return err
	}

	d.data[serviceInstance.InstanceId] = &serviceInstance
	return nil
}

func (d *inMemoryDb) FindServiceInstance(serviceInstanceId string) (*ServiceInstance, error) {
	log.Infof("FindServiceInstance: %s", serviceInstanceId)
	return d.data[serviceInstanceId], nil
}

func (d *inMemoryDb) DeleteServiceInstance(serviceInstanceId string) (*ServiceInstance, error) {
	log.Infof("DeleteServiceInstance: %s", serviceInstanceId)
	serviceInstance := d.data[serviceInstanceId]
	if serviceInstance == nil {
		log.Debugf("No record with key %s found", serviceInstanceId)
	}

	delete(d.data, serviceInstanceId)

	return serviceInstance, nil
}
