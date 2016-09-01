package bosh

import (
	"errors"
	"strings"
)

type Details struct {
	StemcellName    string
	DirectorUUID    string
	NetworkNames    []string
	Vmtype          string
	BoshDirectorUrl string

	PeerDataDir   string
	DockerDataDir string
}

func (b *Details) Validate() error {
	if b.StemcellName == "" {
		return errors.New("StemcellName cannot be empty")
	}
	if b.DirectorUUID == "" {
		return errors.New("DirectorUUID cannot be empty")
	}
	if b.Vmtype == "" {
		return errors.New("Vmtype cannot be empty")
	}
	for _, networkName := range b.NetworkNames {
		if networkName == "" {
			return errors.New("Invalid network name in the list")
		}
	}
	if b.BoshDirectorUrl == "" {
		return errors.New("BoshDirectorUrl cannot be empty")
	}
	if b.PeerDataDir == "" {
		return errors.New("PeerDataDir cannot be empty")
	}
	if b.DockerDataDir == "" {
		return errors.New("DockerDataDir cannot be empty")
	}
	return nil
}

func NewDetails(stemcellName, uuid, vmType, networkNames, boshDirectorUrl, peerDataDir, dockerDataDir string) *Details {
	return &Details{
		StemcellName:    stemcellName,
		DirectorUUID:    uuid,
		Vmtype:          vmType,
		NetworkNames:    strings.Split(strings.Replace(networkNames, " ", "", -1), ","),
		BoshDirectorUrl: boshDirectorUrl,
		PeerDataDir:     peerDataDir,
		DockerDataDir:   dockerDataDir,
	}
}
