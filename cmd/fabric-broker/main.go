package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/atulkc/fabric-service-broker/db"
	"github.com/atulkc/fabric-service-broker/handlers"
	"github.com/atulkc/fabric-service-broker/schema"
	"github.com/atulkc/fabric-service-broker/util"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("fabric-sb")

const (
	defaultScheme       = "https"
	defaultBoshUsername = "admin"
	defaultBoshPassword = "admin"
	defaultBoshAddress  = "192.168.50.4"
	defaultBoshPort     = 25555
	defaultPort         = "8999"
)

var defaultBoshDirectorUrl = fmt.Sprintf("%s://%s:%s@%s:%d", defaultScheme, defaultBoshUsername, defaultBoshPassword, defaultBoshAddress, defaultBoshPort)

var boshDirectorUrl = flag.String(
	"boshDirectorUrl",
	defaultBoshDirectorUrl,
	"Url for BOSH director in format scheme://username:password@ip:port",
)

var boshStemcellName = flag.String(
	"boshStemcellName",
	os.Getenv("BOSH_STEMCELL"),
	"Url for BOSH director in format scheme://username:password@ip:port",
)

var boshDirectorUuid = flag.String(
	"boshDirectorUuid",
	os.Getenv("BOSH_UUID"),
	"BOSH director UUID",
)

var boshVmType = flag.String(
	"boshVmType",
	os.Getenv("BOSH_VM_TYPE"),
	"Vm type defined in cloud config that should be used for peer job",
)

var boshNetworks = flag.String(
	"boshNetworks",
	os.Getenv("BOSH_NETWORK_NAMES"),
	"Comma separated list of network names configured in cloud config",
)

func main() {
	flag.Parse()
	log.Debug("Starting fabric service broker")

	boshDetails := getBoshDetails()
	err := boshDetails.Validate()
	if err != nil {
		log.Error("Environment not setup for bosh director use", err)
		os.Exit(1)
	}

	repo := db.GetInMemoryDB()
	boshClient := util.NewBoshHttpClient(boshDetails)
	slHandler := handlers.NewServiceLifecycleHandler(repo, repo, boshClient, boshDetails)

	r := mux.NewRouter()
	r.HandleFunc("/v2/catalog", handlers.CatalogHandler)
	r.HandleFunc("/v2/service_instances/{instanceId}", slHandler.Provision).Methods("PUT")
	r.HandleFunc("/v2/service_instances/{instanceId}", slHandler.Deprovision).Methods("DELETE")
	r.HandleFunc("/v2/service_instances/{instanceId}/last_operation", slHandler.LastOperation)
	r.HandleFunc("/v2/service_instances/{instanceId}/service_bindings/{bindingId}", slHandler.Bind)

	var port string
	port = os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	log.Debugf("Listening on port: %s", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), r)
}

func getBoshDetails() *schema.BoshDetails {
	log.Info("Getting Bosh details from environment")
	return schema.NewBoshDetails(
		*boshStemcellName,
		*boshDirectorUuid,
		*boshVmType,
		*boshNetworks,
		*boshDirectorUrl,
	)
}
