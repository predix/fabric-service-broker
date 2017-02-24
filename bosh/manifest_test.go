package bosh_test

import (
	"strings"
	"testing"

	"github.com/predix/fabric-service-broker/bosh"

	. "gopkg.in/go-playground/assert.v1"
)

const (
	deploymentName = "mydeployment"
	networkName    = "mynet"
)

var boshDetails = bosh.NewDetails(boshStemcell, boshUuid, vmType, networkNames, directorUrl, peerDataDir, dockerDataDir)

func TestNewManifest(t *testing.T) {
	manifest, err := bosh.NewManifest(deploymentName, networkName, false, boshDetails)

	stemcell := bosh.Stemcell{
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
	Equal(t, manifest.Properties.Peer.Core.DataPath, peerDataDir)
	Equal(t, manifest.Properties.Docker.Store.Dir, dockerDataDir)
}

func TestNewManifestPermissioned(t *testing.T) {
	manifest, err := bosh.NewManifest(deploymentName, networkName, true, boshDetails)

	stemcell := bosh.Stemcell{
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
	Equal(t, manifest.Properties.Peer.Core.DataPath, peerDataDir)
	Equal(t, manifest.Properties.Docker.Store.Dir, dockerDataDir)
	Equal(t, len(manifest.Properties.MemberService.Clients), 7)
	user := bosh.BlockchainUser{
		Name:            "lukas",
		Secret:          "NPKYL39uKbkj",
		Affiliation:     "bank_a",
		AffiliationRole: "00001",
	}
	found := false
	for _, client := range manifest.Properties.MemberService.Clients {
		if client == user {
			found = true
			break
		}
	}
	Equal(t, found, true)
}

func TestManifestToString(t *testing.T) {
	manifest, err := bosh.NewManifest(deploymentName, networkName, false, boshDetails)

	Equal(t, err, nil)
	NotEqual(t, manifest, nil)

	manifest.Name = "test-deployment-name"

	Equal(t, strings.Contains(manifest.String(), "name: test-deployment-name"), true)
	Equal(t, strings.Contains(manifest.String(), "name: hyperledger-fabric"), false)
	Equal(t, strings.Contains(manifest.String(), "plugin: pbft"), true)
}
