package models_test

import (
	"testing"

	dbmodels "github.com/predix/fabric-service-broker/db/models"

	. "gopkg.in/go-playground/assert.v1"
)

const (
	serviceInstanceId   = "instanceId"
	deploymentName      = "deploymentName"
	networkName         = "net1"
	blockChainNetworkId = "blockNetworkId"
	serviceId           = "service-id"
	planId              = "plan-id"
	orgGuid             = "org-guid"
	spaceGuid           = "space-guid"
)

func TestServiceInstance_Validate(t *testing.T) {
	serviceInstance := &dbmodels.ServiceInstance{
		BaseModel:           dbmodels.BaseModel{Id: serviceInstanceId},
		DeploymentName:      deploymentName,
		NetworkName:         networkName,
		BlockchainNetworkId: blockChainNetworkId,
		ServiceId:           serviceId,
		PlanId:              planId,
		OrganizationGuid:    orgGuid,
		SpaceGuid:           spaceGuid,
	}
	err := serviceInstance.Validate()
	Equal(t, err, nil)
}
