package rest_models

type ServiceProvisionRequest struct {
	OrganizationGuid string `json:"organization_guid"`
	PlanId           string `json:"plan_id"`
	ServiceId        string `json:"service_id"`
	SpaceGuid        string `json:"space_guid"`
}
