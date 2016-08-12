package models_test

import (
	"testing"

	dbmodels "github.com/atulkc/fabric-service-broker/db/models"

	. "gopkg.in/go-playground/assert.v1"
)

const (
	serviceInstanceId   = "instanceId"
	deploymentName      = "deploymentName"
	networkName         = "net1"
	blockChainNetworkId = "blockNetworkId"
)

func TestServiceInstance_Validate(t *testing.T) {
	serviceInstance := &dbmodels.ServiceInstance{
		BaseModel:           dbmodels.BaseModel{Id: serviceInstanceId},
		DeploymentName:      deploymentName,
		NetworkName:         networkName,
		BlockchainNetworkId: blockChainNetworkId,
	}
	err := serviceInstance.Validate()
	Equal(t, err, nil)
}
