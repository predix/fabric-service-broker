package rest_models

type BindCredentials struct {
	Credentials BlockChainCredentials `json:"credentials"`
}

type BlockChainCredentials struct {
	PeerEndpoints []string `json:"peers"`
}
