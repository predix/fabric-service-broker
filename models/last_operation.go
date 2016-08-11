package models

const (
	StateInProgress = "in progress"
	StateSucceeded  = "succeeded"
	StateFailed     = "failed"

	OpProvision   = "provision"
	OpDeprovision = "deprovision"
)

type LastOperationResponse struct {
	State       string `json:"state"`
	Description string `json:"description"`
}

func GetLastOperationResponse(operation, boshState string) LastOperationResponse {
	lastOperation := LastOperationResponse{}
	switch boshState {
	case BoshStateProcessing:
		fallthrough
	case BoshStateQueued:
		lastOperation.State = StateInProgress
		if operation == OpProvision {
			lastOperation.Description = "Still working to get that block chain deployed"
		} else {
			lastOperation.Description = "Still working to delete that block chain"
		}
	case BoshStateDone:
		lastOperation.State = StateSucceeded
		if operation == OpProvision {
			lastOperation.Description = "Yipee, block chain is deployed"
		} else {
			lastOperation.Description = "Block chain gone :( Please come back and create another one"
		}
	default:
		lastOperation.State = StateFailed
		if operation == OpProvision {
			lastOperation.Description = "Ooops, could not deploy block chain"
		} else {
			lastOperation.Description = "No we could not delete the block chain..."
		}
	}
	return lastOperation
}
