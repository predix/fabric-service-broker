package schema_test

import (
	"testing"

	"github.com/atulkc/fabric-service-broker/schema"

	. "gopkg.in/go-playground/assert.v1"
)

func TestGetLastOperationResponse_Processing(t *testing.T) {
	lastOperationResponse := schema.GetLastOperationResponse(schema.OpProvision, schema.BoshStateProcessing)
	Equal(t, lastOperationResponse.State, schema.StateInProgress)
}

func TestGetLastOperationResponse_Queued(t *testing.T) {
	lastOperationResponse := schema.GetLastOperationResponse(schema.OpProvision, schema.BoshStateQueued)
	Equal(t, lastOperationResponse.State, schema.StateInProgress)
}

func TestGetLastOperationResponse_Succeeded(t *testing.T) {
	lastOperationResponse := schema.GetLastOperationResponse(schema.OpProvision, schema.BoshStateDone)
	Equal(t, lastOperationResponse.State, schema.StateSucceeded)
}

func TestGetLastOperationResponse_Failed(t *testing.T) {
	lastOperationResponse := schema.GetLastOperationResponse(schema.OpProvision, "failed")
	Equal(t, lastOperationResponse.State, schema.StateFailed)
}
