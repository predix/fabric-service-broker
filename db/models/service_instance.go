package models

import "errors"

type ServiceInstance struct {
	BaseModel
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
	return nil
}
