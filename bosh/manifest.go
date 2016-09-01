package bosh

import (
	"strings"

	"gopkg.in/yaml.v2"
)

const permissionlessManifest = `
---
name: GIVE-ME-A-NAME
director_uuid: REPLACE-ME
stemcells:
- alias: default
  name: USE-IAAS-SPECIFIC-STEMCELL
  version: latest
releases:
- name: fabric-release
  version: latest
update:
  canaries: 1
  canary_watch_time: 5000-120000
  max_in_flight: 3
  serial: false
  update_watch_time: 5000-120000
jobs:
- instances: 4
  azs: [z1, z2]
  name: peer
  networks:
  - name: peer
  persistent_disk: 10000
  vm_type: small
  stemcell: default
  templates:
  - name: peer
    release: fabric-release
  - name: docker
    release: fabric-release
properties:
  peer:
    network:
      id: GENERATED
    consensus:
      plugin: pbft
    core:
      data_path: /var/vcap/store/hyperledger/production
  docker:
    store:
      dir: /var/vcap/store/docker
`

const permissionedManifest = `
---
name: GIVE-ME-A-NAME
director_uuid: REPLACE-ME
stemcells:
- alias: default
  name: USE-IAAS-SPECIFIC-STEMCELL
  version: latest
releases:
- name: fabric-release
  version: latest
update:
  canaries: 1
  canary_watch_time: 5000-120000
  max_in_flight: 3
  serial: false
  update_watch_time: 5000-120000
jobs:
- instances: 1
  azs: [z1, z2]
  name: membersrvc
  networks:
  - name: peer
  persistent_disk: 10000
  vm_type: small
  stemcell: default
  templates:
  - name: member_service
    release: fabric-release
- instances: 4
  azs: [z1, z2]
  name: peer
  networks:
  - name: peer
  persistent_disk: 10000
  vm_type: small
  stemcell: default
  templates:
  - name: peer
    release: fabric-release
  - name: docker
    release: fabric-release
properties:
  peer:
    network:
      id: GENERATED
    consensus:
      plugin: pbft
    security:
      enabled: true
    core:
      data_path: /var/vcap/store/hyperledger/production
  docker:
    store:
      dir: /var/vcap/store/docker
  membersrvc:
    affiliations:
      banks_and_institutions:
        banks:
            - bank_a
            - bank_b
            - bank_c
        institutions:
            - institution_a
            - institution_b
    clients:
    - name: admin
      secret: Xurw3yU9zI0l
      affiliation: institution_a 00001
      affiliation_role:
      metadata: '''{"registrar":{"roles":["client","peer","validator","auditor"],"delegateRoles":["client"]}}'''
    - name: WebAppAdmin
      secret: DJY27pEnl16d
      affiliation: institution_a 00002
      affiliation_role:
      metadata: '''{"registrar":{"roles":["client"]}}'''
    - name: lukas
      secret: NPKYL39uKbkj
      affiliation: bank_a
      affiliation_role: 00001
      metadata:
    - name: system_chaincode_invoker
      secret: DRJ20pEql15a
      affiliation: institution_a
      affiliation_role: 00002
      metadata:
    - name: diego
      secret: DRJ23pEQl16a
      affiliation: institution_a
      affiliation_role: 00003
      metadata:
    - name: binhn
      secret: 7avZQLwcUe9q
      affiliation: institution_a
      affiliation_role: 00005
      metadata:
    - name: jim
      secret: 6avZQLwcUe9b
      affiliation: bank_a
      affiliation_role: 00004
      metadata:
    validators:
    - name: vp0
      secret: f3489fy98ghf
    - name: vp1
      secret: MwYpmSRjupbT
    - name: vp2
      secret: 5wgHK9qqYaPy
    - name: vp3
      secret: vQelbRvja7cJ
    non_validators:
    - name: nvp0
      secret: iywrPBDEPl0K
      affiliation: bank_a
      affiliation_role: 00006
    - name: nvp1
      secret: DcYXuRSocuqd
      affiliation: institution_a
      affiliation_role: 00007
    auditors:
`

type Manifest struct {
	Name         string     `yaml:"name"`
	DirectorUuid string     `yaml:"director_uuid"`
	Stemcells    Stemcells  `yaml:"stemcells"`
	Releases     Releases   `yaml:"releases"`
	Update       Update     `yaml:"update"`
	Jobs         Jobs       `yaml:"jobs"`
	Properties   Properties `yaml:"properties"`
}

type Stemcells []Stemcell

type Stemcell struct {
	Alias   string `yaml:"alias"`
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type Releases []Release

type Release struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type Update struct {
	Canaries        uint   `yaml:"canaries"`
	CanaryWatchTime string `yaml:"canary_watch_time"`
	MaxInFlight     uint   `yaml:"max_in_flight"`
	Serial          bool   `yaml:"serial"`
	UpdateWatchTime string `yaml:"update_watch_time"`
}

type Jobs []Job

type Job struct {
	Instances      uint                `yaml:"instances"`
	AZs            []string            `yaml:"azs"`
	Name           string              `yaml:"name"`
	Networks       []map[string]string `yaml:"networks"`
	PersistentDisk uint                `yaml:"persistent_disk"`
	VmType         string              `yaml:"vm_type"`
	Stemcell       string              `yaml:"stemcell"`
	Templates      []map[string]string `yaml:"templates"`
}

type SecuritySettings struct {
	Enabled bool `yaml:"enabled"`
}

type CoreSettings struct {
	DataPath string `yaml:"data_path"`
}

type PeerProperties struct {
	Network   map[string]string `yaml:"network"`
	Consensus map[string]string `yaml:"consensus"`
	Security  SecuritySettings  `yaml:"security,omitempty"`
	Core      CoreSettings      `yaml:"core,omitempty"`
}

type BlockchainUser struct {
	Name            string `yaml:"name"`
	Secret          string `yaml:"secret"`
	Affiliation     string `yaml:"affiliation,omitempty"`
	AffiliationRole string `yaml:"affiliation_role,omitempty"`
	MetaData        string `yaml:"metadata,omitempty"`
}

type MemberServiceProperties struct {
	Affiliations  map[string]interface{} `yaml:"affiliations"`
	Clients       []BlockchainUser       `yaml:"clients"`
	Validators    []BlockchainUser       `yaml:"validators"`
	NonValidators []BlockchainUser       `yaml:"non_validators"`
	Auditors      []BlockchainUser       `yaml:"auditors"`
}

type StoreSettings struct {
	Dir string `yaml:"dir"`
}

type DockerProperties struct {
	Store StoreSettings `yaml:"store"`
}

type Properties struct {
	Peer          PeerProperties          `yaml:"peer"`
	Docker        DockerProperties        `yaml:"docker,omitempty"`
	MemberService MemberServiceProperties `yaml:"membersrvc,omitempty"`
}

func NewManifest(deploymentName, networkName string, permissioned bool, details *Details) (*Manifest, error) {
	manifest := Manifest{}

	rawManifest := permissionlessManifest
	if permissioned {
		rawManifest = permissionedManifest
	}

	err := yaml.Unmarshal([]byte(rawManifest), &manifest)
	if err != nil {
		log.Error("Error unmarshalling manifest file", err)
		return nil, err
	}

	manifest.Name = deploymentName
	manifest.Properties.Peer.Network["id"] = strings.ToLower(deploymentName)
	manifest.Jobs[0].Networks[0]["name"] = networkName
	manifest.Jobs[0].VmType = details.Vmtype
	if permissioned {
		manifest.Jobs[1].Networks[0]["name"] = networkName
		manifest.Jobs[1].VmType = details.Vmtype
	}
	manifest.DirectorUuid = details.DirectorUUID
	manifest.Stemcells[0].Name = details.StemcellName
	manifest.Properties.Peer.Core.DataPath = details.PeerDataDir
	manifest.Properties.Docker.Store.Dir = details.DockerDataDir

	return &manifest, nil
}

func (m *Manifest) String() string {
	d, err := yaml.Marshal(m)
	if err != nil {
		log.Error("Error marshalling manifest", err)
		return ""
	}

	return string(d)
}
