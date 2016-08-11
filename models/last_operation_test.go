package models_test

import (
	"testing"

	"github.com/atulkc/fabric-service-broker/models"

	. "gopkg.in/go-playground/assert.v1"
)

func TestGetLastOperationResponse_Processing(t *testing.T) {
	lastOperationResponse := models.GetLastOperationResponse(models.OpProvision, models.BoshStateProcessing)
	Equal(t, lastOperationResponse.State, models.StateInProgress)
}

func TestGetLastOperationResponse_Queued(t *testing.T) {
	lastOperationResponse := models.GetLastOperationResponse(models.OpProvision, models.BoshStateQueued)
	Equal(t, lastOperationResponse.State, models.StateInProgress)
}

func TestGetLastOperationResponse_Succeeded(t *testing.T) {
	lastOperationResponse := models.GetLastOperationResponse(models.OpProvision, models.BoshStateDone)
	Equal(t, lastOperationResponse.State, models.StateSucceeded)
}

func TestGetLastOperationResponse_Failed(t *testing.T) {
	lastOperationResponse := models.GetLastOperationResponse(models.OpProvision, "failed")
	Equal(t, lastOperationResponse.State, models.StateFailed)
}
