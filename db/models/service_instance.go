package models

import "errors"

type ServiceInstance struct {
	BaseModel
	ServiceId           string
	PlanId              string
	OrganizationGuid    string
	SpaceGuid           string
	DeploymentName      string
	NetworkName         string
	BlockchainNetworkId string
	ProvisionTaskId     string
	DeprovisionTaskId   string
}

func (s ServiceInstance) Validate() error {
	if s.Id == "" {
		return errors.New("Id cannot be empty")
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
	if s.ServiceId == "" {
		return errors.New("ServiceId cannot be empty")
	}
	if s.PlanId == "" {
		return errors.New("PlanId cannot be empty")
	}
	if s.OrganizationGuid == "" {
		return errors.New("OrganizationGuid cannot be empty")
	}
	if s.SpaceGuid == "" {
		return errors.New("SpaceGuid cannot be empty")
	}
	return nil
}
