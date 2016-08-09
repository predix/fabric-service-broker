package db_test

import (
	"testing"

	"github.com/atulkc/fabric-service-broker/db"

	. "gopkg.in/go-playground/assert.v1"
)

const (
	serviceInstanceId   = "instanceId"
	deploymentName      = "deploymentName"
	networkName         = "net1"
	blockChainNetworkId = "blockNetworkId"
)

func TestServiceInstance_Validate(t *testing.T) {
	serviceInstance := &db.ServiceInstance{
		InstanceId:          serviceInstanceId,
		DeploymentName:      deploymentName,
		NetworkName:         networkName,
		BlockchainNetworkId: blockChainNetworkId,
	}
	err := serviceInstance.Validate()
	Equal(t, err, nil)
}
