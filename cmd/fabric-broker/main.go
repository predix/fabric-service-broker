package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/atulkc/fabric-service-broker/handlers"
	"github.com/atulkc/fabric-service-broker/schema"
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

func main() {
	flag.Parse()
	log.Debug("Starting fabric service broker")
	if *boshDirectorUrl == "" {
		log.Fatal("No BOSH director URL provided")
		os.Exit(1)
	}

	r := mux.NewRouter()
	boshDetails := getBoshDetails()
	err := boshDetails.Validate()
	if err != nil {
		log.Error("Environment not setup for bosh director use", err)
		os.Exit(2)
	}
	slHandler := handlers.NewServiceLlifecycleHandler()
	r.HandleFunc("/v2/catalog", handlers.CatalogHandler)
	r.HandleFunc("/v2/service_instances/{instanceId}", slHandler.Provision).Methods("PUT")
	r.HandleFunc("/v2/service_instances/{instanceId}", slHandler.Deprovision).Methods("DELETE")
	r.HandleFunc("/v2/service_instances/{instanceId}/last_operation", handlers.LastOperationHandler)

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
		os.Getenv("BOSH_STEMCELL"),
		os.Getenv("BOSH_UUID"),
		os.Getenv("BOSH_VM_TYPE"),
		os.Getenv("BOSH_NETWORK_NAMES"),
	)
}
