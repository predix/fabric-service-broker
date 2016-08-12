package rest_models_test

import (
	"testing"

	"github.com/atulkc/fabric-service-broker/bosh"
	"github.com/atulkc/fabric-service-broker/rest_models"

	. "gopkg.in/go-playground/assert.v1"
)

func TestGetLastOperationResponse_Processing(t *testing.T) {
	lastOperationResponse := rest_models.GetLastOperationResponse(rest_models.OpProvision, bosh.BoshStateProcessing)
	Equal(t, lastOperationResponse.State, rest_models.StateInProgress)
}

func TestGetLastOperationResponse_Queued(t *testing.T) {
	lastOperationResponse := rest_models.GetLastOperationResponse(rest_models.OpProvision, bosh.BoshStateQueued)
	Equal(t, lastOperationResponse.State, rest_models.StateInProgress)
}

func TestGetLastOperationResponse_Succeeded(t *testing.T) {
	lastOperationResponse := rest_models.GetLastOperationResponse(rest_models.OpProvision, bosh.BoshStateDone)
	Equal(t, lastOperationResponse.State, rest_models.StateSucceeded)
}

func TestGetLastOperationResponse_Failed(t *testing.T) {
	lastOperationResponse := rest_models.GetLastOperationResponse(rest_models.OpProvision, "failed")
	Equal(t, lastOperationResponse.State, rest_models.StateFailed)
}
