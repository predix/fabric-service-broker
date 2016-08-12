package inmemory

import (
	dbmodels "github.com/atulkc/fabric-service-broker/db/models"
	"github.com/op/go-logging"
)

type inMemoryDb struct {
	serviceInstanceRepo       map[string]dbmodels.ServiceInstance
	serviceBindingRepo        map[string]dbmodels.ServiceBinding
	serviceInstanceBindingMap map[string]dbmodels.ServiceBindings
}

var log = logging.MustGetLogger("inmemory")

var inMemoryDbInstance *inMemoryDb

func init() {
	inMemoryDbInstance = &inMemoryDb{
		serviceInstanceRepo:       make(map[string]dbmodels.ServiceInstance),
		serviceBindingRepo:        make(map[string]dbmodels.ServiceBinding),
		serviceInstanceBindingMap: make(map[string]dbmodels.ServiceBindings),
	}
}

func Get() *inMemoryDb {
	return inMemoryDbInstance
}

func (d *inMemoryDb) UpsertServiceInstance(serviceInstance dbmodels.ServiceInstance) error {
	log.Infof("UpsertServiceInstance: %s", serviceInstance.Id)
	log.Debugf("Body: %#v", serviceInstance)

	err := serviceInstance.Validate()
	if err != nil {
		return err
	}

	d.serviceInstanceRepo[serviceInstance.Id] = serviceInstance
	return nil
}

func (d *inMemoryDb) FindServiceInstance(serviceInstanceId string) (*dbmodels.ServiceInstance, error) {
	log.Infof("FindServiceInstance: %s", serviceInstanceId)
	serviceInstance, found := d.serviceInstanceRepo[serviceInstanceId]
	if !found {
		log.Debugf("No record with key %s found", serviceInstanceId)
		return nil, nil
	}

	return &serviceInstance, nil
}

func (d *inMemoryDb) ListServiceInstances() ([]dbmodels.ServiceInstance, error) {
	log.Infof("ListServiceInstances")

	list := make([]dbmodels.ServiceInstance, len(d.serviceInstanceRepo))
	for _, serviceInstance := range d.serviceInstanceRepo {
		list = append(list, serviceInstance)
	}

	return list, nil
}

func (d *inMemoryDb) DeleteServiceInstance(serviceInstanceId string) (*dbmodels.ServiceInstance, error) {
	log.Infof("DeleteServiceInstance: %s", serviceInstanceId)
	serviceInstance, found := d.serviceInstanceRepo[serviceInstanceId]
	if !found {
		log.Debugf("No record with key %s found", serviceInstanceId)
	}

	delete(d.serviceInstanceRepo, serviceInstanceId)
	delete(d.serviceInstanceBindingMap, serviceInstanceId)

	return &serviceInstance, nil
}

func (d *inMemoryDb) AssociatedServiceBindings(instanceId string) (dbmodels.ServiceBindings, error) {
	log.Infof("AssociatedServiceBindings")
	return d.serviceInstanceBindingMap[instanceId], nil
}

func (d *inMemoryDb) UpsertServiceBinding(serviceBinding dbmodels.ServiceBinding) error {
	log.Infof("UpsertServiceBinding: %s", serviceBinding.Id)
	log.Debugf("Body: %#v", serviceBinding)

	err := serviceBinding.Validate()
	if err != nil {
		return err
	}

	d.serviceBindingRepo[serviceBinding.Id] = serviceBinding
	bindings, found := d.serviceInstanceBindingMap[serviceBinding.ServiceInstanceId]
	if !found {
		bindings = dbmodels.ServiceBindings{}
	}
	bindings = append(bindings, serviceBinding)
	d.serviceInstanceBindingMap[serviceBinding.ServiceInstanceId] = bindings
	return nil
}

func (d *inMemoryDb) FindServiceBinding(bindingId string) (*dbmodels.ServiceBinding, error) {
	log.Infof("FindServiceBinding: %s", bindingId)
	serviceBinding, found := d.serviceBindingRepo[bindingId]
	if !found {
		log.Debugf("No record with key %s found", bindingId)
		return nil, nil
	}

	return &serviceBinding, nil
}

func (d *inMemoryDb) DeleteServiceBinding(bindingId string) (*dbmodels.ServiceBinding, error) {
	log.Infof("DeleteServiceBinding: %s", bindingId)
	serviceBinding, found := d.serviceBindingRepo[bindingId]
	if !found {
		log.Debugf("No record with key %s found", bindingId)
	}

	delete(d.serviceBindingRepo, bindingId)

	bindings, found := d.serviceInstanceBindingMap[serviceBinding.ServiceInstanceId]
	if found {
		newBindings := dbmodels.ServiceBindings{}
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
