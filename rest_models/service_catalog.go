package rest_models

const (
	DefaultServiceId                   = "05FC7A18-5B52-4701-A475-5995B79DF2AD"
	PermissionlessPlanId               = "15175506-D9F6-4CD8-AA1E-8F0AAFB99C07"
	PermissionedPlanId                 = "4D64F255-927B-4807-A358-15CF06EC687B"
	SharedPermissionedPlanId           = "7C1C3178-7551-11E6-8B77-86F30CA893D3"
	SharedPermissionlessPlanId         = "27083A54-76ED-11E6-8B77-86F30CA893D3"
)

type ServiceCatalog struct {
	Services Services `json:"services"`
}

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
				Id:          PermissionlessPlanId,
				Name:        "permissionless",
				Description: "Spins up 4 validating nodes in pbft based block chain",
				Free:        true,
				MetaData: PlanMetaData{
					Name:        "permissionless",
					DisplayName: "Free plan",
					Description: "Dedicated 4 nodes permissionless block chain cluster",
					Costs:       []string{"free"},
				},
			},
			Plan{
				Id:          PermissionedPlanId,
				Name:        "permissioned",
				Description: "Spins up 4 validating nodes in pbft based block chain and membership service. This is permissioned block chain.",
				Free:        true,
				MetaData: PlanMetaData{
					Name:        "permissioned",
					DisplayName: "Free plan",
					Description: "Dedicated 4 nodes permissioned block chain cluster",
					Costs:       []string{"free"},
				},
			},
			Plan{
				Id:          SharedPermissionedPlanId,
				Name:        "shared",
				Description: "Shared 4 node permissioned block chain cluster, all members reuse the one-time commissioned block chain.",
				Free:        true,
				MetaData: PlanMetaData{
					Name:        "shared",
					DisplayName: "Free plan",
					Description: "shared 4 nodes permissioned block chain cluster",
					Costs:       []string{"free"},
				},
			},
			Plan{
				Id:          SharedPermissionlessPlanId,
				Name:        "shared",
				Description: "Shared 4 node permissionless block chain cluster, all members reuse the one-time commissioned block chain.",
				Free:        true,
				MetaData: PlanMetaData{
					Name:        "shared",
					DisplayName: "Free plan",
					Description: "shared 4 nodes permissionless block chain cluster",
					Costs:       []string{"free"},
				},
			},
		},
	}
}
