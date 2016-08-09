package schema

import (
	"errors"
	"os"
	"strings"
)

type boshDetails struct {
	StemcellName string
	DirectorUUID string
	NetworkNames []string
	Vmtype       string
}

func (b *boshDetails) validate() error {
	if b.StemcellName == "" {
		return errors.New("StemcellName cannot be empty")
	}
	if b.DirectorUUID == "" {
		return errors.New("DirectorUUID cannot be empty")
	}
	if b.Vmtype == "" {
		return errors.New("Vmtype cannot be empty")
	}
	if len(b.NetworkNames) == 0 {
		return errors.New("NetworkNames cannot be empty")
	}
	return nil
}

var instance *boshDetails

func init() {
	instance = &boshDetails{
		StemcellName: os.Getenv("BOSH_STEMCELL"),
		DirectorUUID: os.Getenv("BOSH_UUID"),
		Vmtype:       os.Getenv("BOSH_VM_TYPE"),
		NetworkNames: strings.Split(os.Getenv("BOSH_NETWORK_NAMES"), ","),
	}
}

func GetBoshDetails() *boshDetails {
	return instance
}
