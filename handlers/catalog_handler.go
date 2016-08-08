package handlers

import (
	"net/http"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("handler")

const catalogJson = `
{
  "services": [{
    "name": "hyperledger-fabric",
    "id": "3D7690ED-C611-46F4-9E64-F3D2210EE194",
    "description": "Hyperledger fabric block chain service",
    "tags": ["blockchain"],
    "bindable": true,
    "metadata": {
      "provider": {
        "name": "Hyperledger fabric block chain"
      },
      "listing": {
        "blurb": "Hyperledger fabric",
        "longDescription": "Hyperledger fabric block chain - permissioned block chain"
      },
      "displayName": "Hyperledger service broker"
    },
    "plan_updateable": false,
    "plans": [{
      "name": "basic",
      "id": "15175506-D9F6-4CD8-AA1E-8F0AAFB99C07",
      "description": "Spins up 3 validating nodes in pbft based block chain",
      "metadata": {
        "cost": 99,
        "bullets": [{
          "content": "Dedicated 3 nodes block chain cluster"
        }]
      }
    }
    ]
  }]
}
`

func CatalogHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Serving /v2/catalog")
	w.Write([]byte(catalogJson))
}
