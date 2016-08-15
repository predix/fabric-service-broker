# Fabric service broker
This repo is for service broker used to provision/deprovision hyperledger fabric block chain.

## Running service broker
1. Install and start [Bosh lite](https://github.com/cloudfoundry/bosh-lite)
1. Build and upload [fabric bosh release](https://github.com/atulkc/fabric-release)
1. Update the cloud config on bosh director as described [here](https://github.com/atulkc/fabric-release)
1. Fetch the repo
	```
	go get github.com/atulkc/fabric-service-broker
	```
	This will fetch the source under `$GOPATH/src` directory.
1. Execute following command to run the service broker

	```
	cd $GOPATH/src/github.com/atulkc/fabric-service-broker
	go run cmd/fabric-broker/main.go --boshStemcellName bosh-warden-boshlite-ubuntu-trusty-go_agent --boshDirectorUuid $(bosh status --uuid) --boshVmType small --boshNetworks "peer, peer1,peer2, peer3"
	```

## Testing service broker
Once service broker is up and running as described above execute following curl commands to test it out

### Provision
```
curl -v localhost:8999/v2/service_instances/2A98FB4C-B774-45BD-9D5B-7C427933F812?accepts_incomplete=true -X PUT
```

### Last operation
```
curl  localhost:8999/v2/service_instances/2A98FB4C-B774-45BD-9D5B-7C427933F812/last_operation?operation=<task id>
```
`<task id>` is value of `operation` in response from provision operation.

### Bind
```
curl -v  localhost:8999/v2/service_instances/2A98FB4C-B774-45BD-9D5B-7C427933F812/service_bindings/37E1D618-8EBC-4258-99D8-971E67CAAA64 -X PUT
```

### Unbind
```
curl -v  localhost:8999/v2/service_instances/2A98FB4C-B774-45BD-9D5B-7C427933F812/service_bindings/37E1D618-8EBC-4258-99D8-971E67CAAA64 -X DELETE
```

### Deprovision
```
curl -v localhost:8999/v2/service_instances/2A98FB4C-B774-45BD-9D5B-7C427933F812?accepts_incomplete=true -X DELETE
```