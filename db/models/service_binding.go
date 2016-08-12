package models

import "errors"

type ServiceBindings []ServiceBinding
type ServiceBinding struct {
	BaseModel
	ServiceInstanceId string
}

func (b ServiceBinding) Validate() error {
	if b.ServiceInstanceId == "" {
		return errors.New("ServiceInstanceId cannot be empty")
	}
	if b.Id == "" {
		return errors.New("Id cannot be empty")
	}
	return nil
}
