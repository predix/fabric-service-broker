package schema

const (
	BoshStateProcessing = "processing"
	BoshStateQueued     = "queued"
	BoshStateDone       = "done"
)

type Task struct {
	Id          uint   `json:"id"`
	State       string `json:"state"`
	Description string `json:"description"`
	Result      string `json:"result"`
	User        string `json:"user"`
}
