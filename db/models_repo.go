package db

import dbmodels "github.com/atulkc/fabric-service-broker/db/models"

type ModelsRepo interface {
	UpsertServiceInstance(serviceInstance dbmodels.ServiceInstance) error
	FindServiceInstance(serviceInstanceId string) (*dbmodels.ServiceInstance, error)
	ListServiceInstances() ([]dbmodels.ServiceInstance, error)
	DeleteServiceInstance(serviceInstanceId string) (*dbmodels.ServiceInstance, error)

	UpsertServiceBinding(serviceBinding dbmodels.ServiceBinding) error
	FindServiceBinding(bindingId string) (*dbmodels.ServiceBinding, error)
	DeleteServiceBinding(bindingId string) (*dbmodels.ServiceBinding, error)

	AssociatedServiceBindings(serviceInstanceId string) (dbmodels.ServiceBindings, error)
}
