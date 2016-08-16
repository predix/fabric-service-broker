package rest_models

type ServiceBindingRequest struct {
	PlanId    string `json:"plan_id"`
	ServiceId string `json:"service_id"`
	AppGuid   string `json:"app_guid"`
}
