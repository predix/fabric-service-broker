package db

import dbmodels "github.com/predix/fabric-service-broker/db/models"

type ModelsRepo interface {
	CreateServiceInstance(serviceInstance dbmodels.ServiceInstance) error
	UpdateServiceInstance(serviceInstance dbmodels.ServiceInstance) error
	FindServiceInstance(serviceInstanceId string) (*dbmodels.ServiceInstance, error)
	ListServiceInstances() ([]dbmodels.ServiceInstance, error)
	DeleteServiceInstance(serviceInstanceId string) (*dbmodels.ServiceInstance, error)

	CreateServiceBinding(serviceBinding dbmodels.ServiceBinding) error
	UpdateServiceBinding(serviceBinding dbmodels.ServiceBinding) error
	FindServiceBinding(bindingId string) (*dbmodels.ServiceBinding, error)
	DeleteServiceBinding(bindingId string) (*dbmodels.ServiceBinding, error)

	AssociatedServiceBindings(serviceInstanceId string) (dbmodels.ServiceBindings, error)
}
