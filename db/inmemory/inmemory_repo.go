package inmemory

import (
	"github.com/atulkc/fabric-service-broker/db/models"
	"github.com/op/go-logging"
)

type inMemoryDb struct {
	serviceInstanceRepo       map[string]models.ServiceInstance
	serviceBindingRepo        map[string]models.ServiceBinding
	serviceInstanceBindingMap map[string]models.ServiceBindings
}

var log = logging.MustGetLogger("inmemory")

var inMemoryDbInstance *inMemoryDb

func init() {
	inMemoryDbInstance = &inMemoryDb{
		serviceInstanceRepo:       make(map[string]models.ServiceInstance),
		serviceBindingRepo:        make(map[string]models.ServiceBinding),
		serviceInstanceBindingMap: make(map[string]models.ServiceBindings),
	}
}

func Get() *inMemoryDb {
	return inMemoryDbInstance
}

func (d *inMemoryDb) CreateServiceInstance(serviceInstance models.ServiceInstance) error {
	log.Infof("CreateServiceInstance: %s", serviceInstance.Id)
	return d.setServiceInstance(serviceInstance)
}

func (d *inMemoryDb) UpdateServiceInstance(serviceInstance models.ServiceInstance) error {
	log.Infof("UpdateServiceInstance: %s", serviceInstance.Id)
	return d.setServiceInstance(serviceInstance)
}

func (d *inMemoryDb) setServiceInstance(serviceInstance models.ServiceInstance) error {
	log.Debugf("Body: %#v", serviceInstance)

	err := serviceInstance.Validate()
	if err != nil {
		return err
	}

	d.serviceInstanceRepo[serviceInstance.Id] = serviceInstance
	return nil
}

func (d *inMemoryDb) FindServiceInstance(serviceInstanceId string) (*models.ServiceInstance, error) {
	log.Infof("FindServiceInstance: %s", serviceInstanceId)
	serviceInstance, found := d.serviceInstanceRepo[serviceInstanceId]
	if !found {
		log.Debugf("No record with key %s found", serviceInstanceId)
		return nil, nil
	}

	return &serviceInstance, nil
}

func (d *inMemoryDb) ListServiceInstances() ([]models.ServiceInstance, error) {
	log.Infof("ListServiceInstances")

	list := make([]models.ServiceInstance, len(d.serviceInstanceRepo))
	for _, serviceInstance := range d.serviceInstanceRepo {
		list = append(list, serviceInstance)
	}

	return list, nil
}

func (d *inMemoryDb) DeleteServiceInstance(serviceInstanceId string) (*models.ServiceInstance, error) {
	log.Infof("DeleteServiceInstance: %s", serviceInstanceId)
	serviceInstance, found := d.serviceInstanceRepo[serviceInstanceId]
	if !found {
		log.Debugf("No record with key %s found", serviceInstanceId)
	}

	delete(d.serviceInstanceRepo, serviceInstanceId)
	delete(d.serviceInstanceBindingMap, serviceInstanceId)

	return &serviceInstance, nil
}

func (d *inMemoryDb) AssociatedServiceBindings(instanceId string) (models.ServiceBindings, error) {
	log.Infof("AssociatedServiceBindings")
	return d.serviceInstanceBindingMap[instanceId], nil
}

func (d *inMemoryDb) CreateServiceBinding(serviceBinding models.ServiceBinding) error {
	log.Infof("CreateServiceBinding: %s", serviceBinding.Id)
	return d.setServiceBinding(serviceBinding)
}

func (d *inMemoryDb) UpdateServiceBinding(serviceBinding models.ServiceBinding) error {
	log.Infof("UpdateServiceBinding: %s", serviceBinding.Id)
	return d.setServiceBinding(serviceBinding)
}

func (d *inMemoryDb) setServiceBinding(serviceBinding models.ServiceBinding) error {
	log.Debugf("Body: %#v", serviceBinding)

	err := serviceBinding.Validate()
	if err != nil {
		return err
	}

	d.serviceBindingRepo[serviceBinding.Id] = serviceBinding
	bindings, found := d.serviceInstanceBindingMap[serviceBinding.ServiceInstanceId]
	if !found {
		bindings = models.ServiceBindings{}
	}
	bindings = append(bindings, serviceBinding)
	d.serviceInstanceBindingMap[serviceBinding.ServiceInstanceId] = bindings
	return nil
}

func (d *inMemoryDb) FindServiceBinding(bindingId string) (*models.ServiceBinding, error) {
	log.Infof("FindServiceBinding: %s", bindingId)
	serviceBinding, found := d.serviceBindingRepo[bindingId]
	if !found {
		log.Debugf("No record with key %s found", bindingId)
		return nil, nil
	}

	return &serviceBinding, nil
}

func (d *inMemoryDb) DeleteServiceBinding(bindingId string) (*models.ServiceBinding, error) {
	log.Infof("DeleteServiceBinding: %s", bindingId)
	serviceBinding, found := d.serviceBindingRepo[bindingId]
	if !found {
		log.Debugf("No record with key %s found", bindingId)
	}

	delete(d.serviceBindingRepo, bindingId)

	bindings, found := d.serviceInstanceBindingMap[serviceBinding.ServiceInstanceId]
	if found {
		newBindings := models.ServiceBindings{}
		for _, binding := range bindings {
			if binding.Id != bindingId {
				newBindings = append(newBindings, binding)
			}
		}
		if len(newBindings) == 0 {
			delete(d.serviceInstanceBindingMap, serviceBinding.ServiceInstanceId)
		} else {
			d.serviceInstanceBindingMap[serviceBinding.ServiceInstanceId] = newBindings
		}
	}

	return &serviceBinding, nil
}
