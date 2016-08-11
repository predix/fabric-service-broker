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

type ServiceBindings []ServiceBinding
type ServiceBinding struct {
	InstanceId string
	BindingId  string
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

func (b ServiceBinding) Validate() error {
	if b.InstanceId == "" {
		return errors.New("InstanceId cannot be empty")
	}
	if b.BindingId == "" {
		return errors.New("BindingId cannot be empty")
	}
	return nil
}

type ServiceInstanceRepo interface {
	Upsert(serviceInstance ServiceInstance) error
	Find(serviceInstanceId string) (*ServiceInstance, error)
	List() ([]ServiceInstance, error)
	Delete(serviceInstanceId string) (*ServiceInstance, error)
}

type ServiceBindingRepo interface {
	UpsertBinding(serviceBinding ServiceBinding) error
	FindBinding(bindingId string) (*ServiceBinding, error)
	DeleteBinding(bindingId string) (*ServiceBinding, error)
}

type inMemoryDb struct {
	serviceInstanceRepo map[string]ServiceInstance
	serviceBindingRepo  map[string]ServiceBinding
}

var log = logging.MustGetLogger("db")

var inMemoryDbInstance *inMemoryDb

func init() {
	inMemoryDbInstance = &inMemoryDb{
		serviceInstanceRepo: make(map[string]ServiceInstance),
		serviceBindingRepo:  make(map[string]ServiceBinding),
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

	d.serviceInstanceRepo[serviceInstance.InstanceId] = serviceInstance
	return nil
}

func (d *inMemoryDb) Find(serviceInstanceId string) (*ServiceInstance, error) {
	log.Infof("FindServiceInstance: %s", serviceInstanceId)
	serviceInstance, found := d.serviceInstanceRepo[serviceInstanceId]
	if !found {
		log.Debugf("No record with key %s found", serviceInstanceId)
		return nil, nil
	}

	return &serviceInstance, nil
}

func (d *inMemoryDb) List() ([]ServiceInstance, error) {
	log.Infof("List")

	list := make([]ServiceInstance, len(d.serviceInstanceRepo))
	for _, serviceInstance := range d.serviceInstanceRepo {
		list = append(list, serviceInstance)
	}

	return list, nil
}

func (d *inMemoryDb) Delete(serviceInstanceId string) (*ServiceInstance, error) {
	log.Infof("DeleteServiceInstance: %s", serviceInstanceId)
	serviceInstance, found := d.serviceInstanceRepo[serviceInstanceId]
	if !found {
		log.Debugf("No record with key %s found", serviceInstanceId)
	}

	delete(d.serviceInstanceRepo, serviceInstanceId)

	return &serviceInstance, nil
}

func (d *inMemoryDb) UpsertBinding(serviceBinding ServiceBinding) error {
	log.Infof("UpsertServiceBinding: %s", serviceBinding.BindingId)
	log.Debugf("Body: %#v", serviceBinding)

	err := serviceBinding.Validate()
	if err != nil {
		return err
	}

	d.serviceBindingRepo[serviceBinding.BindingId] = serviceBinding
	return nil
}

func (d *inMemoryDb) FindBinding(bindingId string) (*ServiceBinding, error) {
	log.Infof("FindServiceBinding: %s", bindingId)
	serviceBinding, found := d.serviceBindingRepo[bindingId]
	if !found {
		log.Debugf("No record with key %s found", bindingId)
		return nil, nil
	}

	return &serviceBinding, nil
}

func (d *inMemoryDb) DeleteBinding(bindingId string) (*ServiceBinding, error) {
	log.Infof("DeleteServiceBinding: %s", bindingId)
	serviceBinding, found := d.serviceBindingRepo[bindingId]
	if !found {
		log.Debugf("No record with key %s found", bindingId)
	}

	delete(d.serviceBindingRepo, bindingId)

	return &serviceBinding, nil
}
