package postgres

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/op/go-logging"
	"github.com/predix/fabric-service-broker/db/models"
)

var log = logging.MustGetLogger("postgres")

// type Credentials struct {
// 	Dbname   string `json:"dbname"`
// 	Hostname string `json:"hostname"`
// 	Password string `json:"password"`
// 	Username string `json:"username"`
// 	Port     int    `json:"port"`
// 	Uri      string `json:"uri"`
// }

// func NewCredentials(uri string) (Credentials, error) {
// }

type postgresDb struct {
	db *gorm.DB
}

// Not a thread safe implementation. It is expected that caller does required
// locking before invoking any methods.
// Could be changed to be thread safe but will be handled in bigger context
// of how concurrency is handled for multiple instances of server.
func New(uri string, migrate bool) (*postgresDb, error) {
	db, err := gorm.Open("postgres", uri)
	if err != nil {
		log.Error("Error connecting to Database", err)
		return nil, err
	}

	if migrate {
		log.Info("Performing auto migration")
		db.AutoMigrate(&models.ServiceInstance{})
		db.AutoMigrate(&models.ServiceBinding{})
	}

	return &postgresDb{
		db: db,
	}, nil
}

func (d *postgresDb) CreateServiceInstance(serviceInstance models.ServiceInstance) error {
	log.Infof("CreateServiceInstance: %s", serviceInstance.Id)
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
		return errors.New(fmt.Sprintf("Service instance: %s already exists", serviceInstance.Id))
	}

	return nil
}

func (d *postgresDb) UpdateServiceInstance(serviceInstance models.ServiceInstance) error {
	log.Infof("UpdateServiceInstance: %s", serviceInstance.Id)
	log.Debugf("Body: %#v", serviceInstance)

	err := serviceInstance.Validate()
	if err != nil {
		return err
	}

	existingInstance, err := d.FindServiceInstance(serviceInstance.Id)
	if err != nil {
		return err
	}

	if existingInstance != nil {
		// update
		err = d.db.Save(&serviceInstance).Error
		if err != nil {
			return err
		}
	} else {
		return errors.New(fmt.Sprintf("Service instance: %s does not exist", serviceInstance.Id))
	}

	return nil
}

func (d *postgresDb) FindServiceInstance(serviceInstanceId string) (*models.ServiceInstance, error) {
	log.Infof("FindServiceInstance: %s", serviceInstanceId)

	existingInstance := models.ServiceInstance{}
	err := d.db.Where("id =?", serviceInstanceId).First(&existingInstance).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if existingInstance.Id != serviceInstanceId {
		log.Debugf("No record with key %s found", serviceInstanceId)
		return nil, nil
	}

	return &existingInstance, nil
}

func (d *postgresDb) ListServiceInstances() ([]models.ServiceInstance, error) {
	log.Infof("ListServiceInstances")
	list := make([]models.ServiceInstance, 0)
	err := d.db.Find(&list).Error
	return list, err
}

func (d *postgresDb) DeleteServiceInstance(serviceInstanceId string) (*models.ServiceInstance, error) {
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

func (d *postgresDb) AssociatedServiceBindings(instanceId string) (models.ServiceBindings, error) {
	log.Infof("AssociatedServiceBindings")

	serviceBindings := models.ServiceBindings{}

	err := d.db.Where("service_instance_id =?", instanceId).Find(&serviceBindings).Error
	return serviceBindings, err
}

func (d *postgresDb) FindServiceBinding(bindingId string) (*models.ServiceBinding, error) {
	log.Infof("FindServiceBinding: %s", bindingId)
	serviceBinding := models.ServiceBinding{}
	err := d.db.Where("id =?", bindingId).First(&serviceBinding).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if serviceBinding.Id != bindingId {
		log.Debugf("No record with key %s found", bindingId)
		return nil, nil
	}

	return &serviceBinding, nil
}

func (d *postgresDb) CreateServiceBinding(serviceBinding models.ServiceBinding) error {
	log.Infof("CreateServiceBinding: %s", serviceBinding.Id)
	log.Debugf("Body: %#v", serviceBinding)

	err := serviceBinding.Validate()
	if err != nil {
		return err
	}

	existingBinding, err := d.FindServiceBinding(serviceBinding.Id)
	if err != nil {
		return err
	}

	if existingBinding == nil {
		// create
		err = d.db.Create(&serviceBinding).Error
		if err != nil {
			return err
		}
	} else {
		return errors.New(fmt.Sprintf("Service binding: %s already exists", serviceBinding.Id))
	}

	return nil
}

func (d *postgresDb) UpdateServiceBinding(serviceBinding models.ServiceBinding) error {
	log.Infof("UpdateServiceBinding: %s", serviceBinding.Id)
	log.Debugf("Body: %#v", serviceBinding)

	err := serviceBinding.Validate()
	if err != nil {
		return err
	}

	existingBinding, err := d.FindServiceBinding(serviceBinding.Id)
	if err != nil {
		return err
	}

	if existingBinding != nil {
		// update
		err = d.db.Save(&serviceBinding).Error
		if err != nil {
			return err
		}
	} else {
		return errors.New(fmt.Sprintf("Service binding: %s does not exist", serviceBinding.Id))
	}

	return nil
}

func (d *postgresDb) DeleteServiceBinding(bindingId string) (*models.ServiceBinding, error) {
	log.Infof("DeleteServiceBinding: %s", bindingId)
	existingBinding, err := d.FindServiceBinding(bindingId)
	if err != nil {
		return nil, err
	}

	if existingBinding != nil {
		err = d.db.Delete(existingBinding).Error
		if err != nil {
			return nil, err
		}
	}

	return existingBinding, nil
}
