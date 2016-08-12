package postgres

import (
	"fmt"

	dbmodels "github.com/atulkc/fabric-service-broker/db/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("postgres")

type postgresDb struct {
	db *gorm.DB
}

func New(host string, port int, dbName, user, secret string, sslDisabled bool) (*postgresDb, error) {
	sslMode := "require"
	if sslDisabled {
		sslMode = "disable"
	}

	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		host, port, dbName, user, secret, sslMode))
	if err != nil {
		log.Error("Error connecting to Database", err)
		return nil, err
	}

	db.AutoMigrate(&dbmodels.ServiceInstance{})
	db.AutoMigrate(&dbmodels.ServiceBinding{})

	return &postgresDb{
		db: db,
	}, nil
}

func (d *postgresDb) UpsertServiceInstance(serviceInstance dbmodels.ServiceInstance) error {
	log.Infof("UpsertServiceInstance: %s", serviceInstance.Id)
	log.Debugf("Body: %#v", serviceInstance)

	err := serviceInstance.Validate()
	if err != nil {
		return err
	}

	existingInstance, err := d.FindServiceInstance(serviceInstance.Id)
	if err != nil {
		return err
	}

	if existingInstance == nil {
		// create
		err = d.db.Create(&serviceInstance).Error
		if err != nil {
			return err
		}
	} else {
		// update
		err = d.db.Save(&serviceInstance).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *postgresDb) FindServiceInstance(serviceInstanceId string) (*dbmodels.ServiceInstance, error) {
	log.Infof("FindServiceInstance: %s", serviceInstanceId)

	existingInstance := dbmodels.ServiceInstance{}
	err := d.db.Where("id =?", serviceInstanceId).First(&existingInstance).Error
	if err != nil {
		return nil, err
	}

	if existingInstance.Id != serviceInstanceId {
		log.Debugf("No record with key %s found", serviceInstanceId)
		return nil, nil
	}

	return &existingInstance, nil
}

func (d *postgresDb) ListServiceInstances() ([]dbmodels.ServiceInstance, error) {
	log.Infof("ListServiceInstances")
	list := make([]dbmodels.ServiceInstance, 0)
	err := d.db.Find(&list).Error
	return list, err
}

func (d *postgresDb) DeleteServiceInstance(serviceInstanceId string) (*dbmodels.ServiceInstance, error) {
	log.Infof("DeleteServiceInstance: %s", serviceInstanceId)

	existingInstance, err := d.FindServiceInstance(serviceInstanceId)
	if err != nil {
		return nil, err
	}

	if existingInstance != nil {
		err = d.db.Delete(existingInstance).Error
		if err != nil {
			return nil, err
		}
	}

	return existingInstance, nil
}

func (d *postgresDb) AssociatedServiceBindings(instanceId string) (dbmodels.ServiceBindings, error) {
	log.Infof("AssociatedServiceBindings")

	serviceBindings := dbmodels.ServiceBindings{}

	err := d.db.Where("service_instance_id =?", instanceId).Find(&serviceBindings).Error
	return serviceBindings, err
}

func (d *postgresDb) UpsertServiceBinding(serviceBinding dbmodels.ServiceBinding) error {
	log.Infof("UpsertServiceBinding: %s", serviceBinding.Id)
	log.Debugf("Body: %#v", serviceBinding)

	err := serviceBinding.Validate()
	if err != nil {
		return err
	}

	d.serviceBindingRepo[serviceBinding.Id] = serviceBinding
	bindings, found := d.serviceInstanceBindingMap[serviceBinding.ServiceInstanceId]
	if !found {
		bindings = dbmodels.ServiceBindings{}
	}
	bindings = append(bindings, serviceBinding)
	d.serviceInstanceBindingMap[serviceBinding.ServiceInstanceId] = bindings
	return nil
}

func (d *postgresDb) FindServiceBinding(bindingId string) (*dbmodels.ServiceBinding, error) {
	log.Infof("FindServiceBinding: %s", bindingId)
	serviceBinding, found := d.serviceBindingRepo[bindingId]
	if !found {
		log.Debugf("No record with key %s found", bindingId)
		return nil, nil
	}

	return &serviceBinding, nil
}

func (d *postgresDb) DeleteServiceBinding(bindingId string) (*dbmodels.ServiceBinding, error) {
	log.Infof("DeleteServiceBinding: %s", bindingId)
	serviceBinding, found := d.serviceBindingRepo[bindingId]
	if !found {
		log.Debugf("No record with key %s found", bindingId)
	}

	delete(d.serviceBindingRepo, bindingId)

	bindings, found := d.serviceInstanceBindingMap[serviceBinding.ServiceInstanceId]
	if found {
		newBindings := dbmodels.ServiceBindings{}
		for _, binding := range bindings {
			if binding.Id != bindingId {
				newBindings = append(newBindings, binding)
			}
		}
		if len(newBindings) == 0 {
			delete(d.serviceInstanceBindingMap, serviceBinding.ServiceInstanceId)
		} else {
			d.serviceInstanceBindingMap[serviceBinding.ServiceInstanceId] = newBindings
		}
	}

	return &serviceBinding, nil
}
