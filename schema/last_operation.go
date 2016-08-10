package schema

const (
	StateInProgress = "in progress"
	StateSucceeded  = "succeeded"
	StateFailed     = "failed"
)

type LastOperationResponse struct {
	State       string `json:"state"`
	Description string `json:"description"`
}

func GetLastOperationResponse(boshState string) LastOperationResponse {
	lastOperation := LastOperationResponse{}
	switch boshState {
	case BoshStateProcessing:
		fallthrough
	case BoshStateQueued:
		lastOperation.State = StateInProgress
		lastOperation.Description = "Still working to get that block chain deployed"
	case BoshStateDone:
		lastOperation.State = StateSucceeded
		lastOperation.Description = "Yipee, block chain is deployed"
	default:
		lastOperation.State = StateFailed
		lastOperation.Description = "Ooops, could not deploy block chain"
	}
	return lastOperation
}
