package schema_test

import (
	"testing"

	"github.com/atulkc/fabric-service-broker/schema"

	. "gopkg.in/go-playground/assert.v1"
)

const (
	boshStemcell = "mystemcell"
	boshUuid     = "uuid-1"
	vmType       = "vmtype"
	networkNames = "net1,net2,net3"
)

func TestNewBoshDetails(t *testing.T) {
	boshDetails := schema.NewBoshDetails(boshStemcell, boshUuid, vmType, networkNames)
	NotEqual(t, boshDetails, nil)
	Equal(t, boshDetails.StemcellName, boshStemcell)
	Equal(t, boshDetails.DirectorUUID, boshUuid)
	Equal(t, boshDetails.Vmtype, vmType)
	Equal(t, len(boshDetails.NetworkNames), 3)
}
