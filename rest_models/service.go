package rest_models

const (
	DefaultServiceId = "05FC7A18-5B52-4701-A475-5995B79DF2AD"
	DefautPlanId     = "15175506-D9F6-4CD8-AA1E-8F0AAFB99C07"
)

type Services []Service

type Service struct {
	Name          string          `json:"name"`
	Id            string          `json:"id"`
	Description   string          `json:"description"`
	Tags          []string        `json:"tags"`
	Bindable      bool            `json:"bindable"`
	MetaData      ServiceMetaData `json:"metadata"`
	PlanUpdatable bool            `json:"plan_updateable"`
	Plans         []Plan          `json:"plans"`
}

type ServiceMetaData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DisplayName string `json:"displayName"`
}

type Plan struct {
	Name        string       `json:"name"`
	Id          string       `json:"id"`
	Description string       `json:"description"`
	MetaData    PlanMetaData `json:"metadata"`
	Free        bool         `json:"free"`
}

type PlanMetaData struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	DisplayName string   `json:"displayName"`
	Costs       []string `json:"costs"`
}

func GetDefaultService() Service {
	return Service{
		Name:        "hyperledger-fabric",
		Id:          DefaultServiceId,
		Description: "Hyperledger fabric block chain service",
		Tags:        []string{"blockchain"},
		Bindable:    true,
		MetaData: ServiceMetaData{
			Name:        "hyperledger-fabric",
			DisplayName: "Hyperledger fabric block chain",
			Description: "Permissioned block chain implementation",
		},
		PlanUpdatable: false,
		Plans: []Plan{
			Plan{
				Id:          DefautPlanId,
				Name:        "basic",
				Description: "Spins up 3 validating nodes in pbft based block chain",
				Free:        true,
				MetaData: PlanMetaData{
					Name:        "basic",
					DisplayName: "Free plan",
					Description: "Dedicated 3 nodes block chain cluster",
					Costs:       []string{"free"},
				},
			},
		},
	}
}
