package schema_test

import (
	"strings"
	"testing"

	"github.com/atulkc/fabric-service-broker/schema"

	. "gopkg.in/go-playground/assert.v1"
)

const (
	deploymentName = "mydeployment"
	networkName    = "mynet"
)

var boshDetails = schema.NewBoshDetails(boshStemcell, boshUuid, vmType, networkNames, directorUrl)

func TestNewManifest(t *testing.T) {
	manifest, err := schema.NewManifest(deploymentName, networkName, boshDetails)

	stemcell := schema.Stemcell{
		Alias:   "default",
		Name:    boshStemcell,
		Version: "latest",
	}

	Equal(t, err, nil)
	NotEqual(t, manifest, nil)

	Equal(t, manifest.Name, deploymentName)
	Equal(t, manifest.DirectorUuid, boshUuid)
	Equal(t, manifest.Stemcells[0], stemcell)
	Equal(t, manifest.Jobs[0].VmType, vmType)
	Equal(t, manifest.Jobs[0].Networks[0], map[string]string{"name": networkName})
	Equal(t, manifest.Properties.Peer.Network, map[string]string{"id": deploymentName})
	Equal(t, manifest.Properties.Peer.Consensus, map[string]string{"plugin": "pbft"})
}

func TestManifestToString(t *testing.T) {
	manifest, err := schema.NewManifest(deploymentName, networkName, boshDetails)

	Equal(t, err, nil)
	NotEqual(t, manifest, nil)

	manifest.Name = "test-deployment-name"

	Equal(t, strings.Contains(manifest.String(), "name: test-deployment-name"), true)
	Equal(t, strings.Contains(manifest.String(), "name: hyperledger-fabric"), false)
	Equal(t, strings.Contains(manifest.String(), "plugin: pbft"), true)
}
