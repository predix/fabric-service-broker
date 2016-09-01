package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/atulkc/fabric-service-broker/bosh"
	"github.com/atulkc/fabric-service-broker/db"
	"github.com/atulkc/fabric-service-broker/db/inmemory"
	"github.com/atulkc/fabric-service-broker/db/postgres"
	"github.com/atulkc/fabric-service-broker/handlers"
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var log = logging.MustGetLogger("fabric-sb")

const (
	defaultScheme        = "https"
	defaultBoshUsername  = "admin"
	defaultBoshPassword  = "admin"
	defaultBoshAddress   = "192.168.50.4"
	defaultBoshPort      = 25555
	defaultPort          = "8999"
	defaultPeerDataDir   = "/var/vcap/data/hyperledger/production"
	defaultDockerDataDir = "/var/vcap/data/docker"
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

var peerDataDir = flag.String(
	"peerDataDir",
	defaultPeerDataDir,
	"Data directory used by peer to store data",
)

var dockerDataDir = flag.String(
	"dockerDataDir",
	defaultDockerDataDir,
	"Data directory used by docker to store data files",
)

var dbUrl = flag.String(
	"dbUrl",
	os.Getenv("DB_CONNECTION_STRING"),
	"Url for DB in DB specific format. E.g. for postgres it will be postgres://__username__:__password__@__hostname__:__port__/__database__",
)

func main() {
	flag.Parse()
	log.Debug("Starting fabric service broker")

	connectionString := ""
	if os.Getenv("VCAP_APPLICATION") != "" {
		appEnv, err := cfenv.Current()
		if err != nil {
			log.Error("Could not read CF App environment", err)
			os.Exit(1)
		}
		log.Debugf("Instance index is :%d", appEnv.Index)
		//TODO: Get connection string from VCAP_SERVICES
	} else {
		log.Info("Not running as CF App")
	}

	var repo db.ModelsRepo
	if connectionString == "" {
		log.Info("Connection string not available from VCAP_SERVICES")
		if *dbUrl == "" {
			log.Info("No db url specified as CLI parameter, using inmemory DB")
			repo = inmemory.Get()
		} else {
			log.Info("DB Url specified as CLI parameter, using postgres repo")
			repo = getPostgresRepo(*dbUrl)
		}
	} else {
		log.Info("Connection string available from VCAP_SERVICES, using postgres repo")
		repo = getPostgresRepo(connectionString)
	}

	boshDetails := getBoshDetails()
	err := boshDetails.Validate()
	if err != nil {
		log.Error("Environment not setup for bosh director use", err)
		os.Exit(2)
	}

	boshClient := bosh.NewBoshHttpClient(boshDetails)
	slHandler := handlers.NewServiceLifecycleHandler(repo, boshClient, boshDetails)

	r := mux.NewRouter()
	r.HandleFunc("/v2/catalog", handlers.CatalogHandler)
	r.HandleFunc("/v2/service_instances/{instanceId}", slHandler.Provision).Methods("PUT")
	r.HandleFunc("/v2/service_instances/{instanceId}", slHandler.Deprovision).Methods("DELETE")
	r.HandleFunc("/v2/service_instances/{instanceId}/last_operation", slHandler.LastOperation)
	r.HandleFunc("/v2/service_instances/{instanceId}/service_bindings/{bindingId}", slHandler.Bind).Methods("PUT")
	r.HandleFunc("/v2/service_instances/{instanceId}/service_bindings/{bindingId}", slHandler.Unbind).Methods("DELETE")

	var port string
	port = os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	log.Debugf("Listening on port: %s", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), r)
}

func getBoshDetails() *bosh.Details {
	log.Info("Getting Bosh details from environment")
	return bosh.NewDetails(
		*boshStemcellName,
		*boshDirectorUuid,
		*boshVmType,
		*boshNetworks,
		*boshDirectorUrl,
		*peerDataDir,
		*dockerDataDir,
	)
}

func getPostgresRepo(uri string) db.ModelsRepo {
	repo, err := postgres.New(*dbUrl, true)
	if err != nil {
		log.Error("Error opening DB connection", err)
		os.Exit(1)
	}
	return repo
}
