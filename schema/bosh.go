package schema

import (
	"errors"
	"strings"
)

type BoshDetails struct {
	StemcellName string
	DirectorUUID string
	NetworkNames []string
	Vmtype       string
}

func (b *BoshDetails) Validate() error {
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

func NewBoshDetails(stemcellName, uuid, vmType, networkNames string) *BoshDetails {
	return &BoshDetails{
		StemcellName: stemcellName,
		DirectorUUID: uuid,
		Vmtype:       vmType,
		NetworkNames: strings.Split(networkNames, ","),
	}
}
