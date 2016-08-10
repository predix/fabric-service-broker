package util

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	sberrors "github.com/atulkc/fabric-service-broker/errors"
	"github.com/atulkc/fabric-service-broker/schema"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("util")

type BoshClient interface {
	CreateDeployment(manifest schema.Manifest) (*schema.Task, error)
	DeleteDeployment(deploymentName string) (*schema.Task, error)
	GetTask(taskId string) (*schema.Task, error)
}

type boshHttpClient struct {
	httpClient  *http.Client
	boshDetails *schema.BoshDetails
}

func newHttpClient(skipTLSVerification bool) *http.Client {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: skipTLSVerification,
	}
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errors.New("No redirects")
		},
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).Dial,
			TLSClientConfig: tlsConfig,
		},
	}
}

func NewBoshHttpClient(boshDetails *schema.BoshDetails) BoshClient {
	return &boshHttpClient{
		boshDetails: boshDetails,
		httpClient:  newHttpClient(true),
	}
}

func (c *boshHttpClient) CreateDeployment(manifest schema.Manifest) (*schema.Task, error) {
	log.Debug("In CreateDeployment")
	body := manifest.String()
	log.Debugf("Manifest for deployment:%s", body)

	url := fmt.Sprintf("%s%s", c.boshDetails.BoshDirectorUrl, "/deployments")
	request, err := http.NewRequest("POST", url, bytes.NewReader([]byte(body)))
	if err != nil {
		log.Error("Error in creating http request", err)
		return nil, errors.New(sberrors.ErrHttpRequest)
	}
	request.Header.Set("Content-Type", "text/yaml")
	log.Debugf("Http request for BOSH director created")

	resp, err := c.httpClient.Do(request)
	if err != nil && !strings.Contains(err.Error(), "No redirects") {
		log.Error("Error in connecting to Bosh", err)
		return nil, errors.New(sberrors.ErrBoshConnect)
	}

	taskId, err := getTaskId(resp)
	if err != nil {
		return nil, err
	}

	log.Infof("Successfully initiated deployment:%s. Task Id is: %s", manifest.Name, taskId)

	return c.GetTask(taskId)
}

func (c *boshHttpClient) DeleteDeployment(deploymentName string) (*schema.Task, error) {
	log.Debug("In DeleteDeployment")
	url := fmt.Sprintf("%s%s%s", c.boshDetails.BoshDirectorUrl, "/deployments/", deploymentName)

	delRequest, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Error("Error in creating http request", err)
		return nil, errors.New(sberrors.ErrHttpRequest)
	}

	resp, err := c.httpClient.Do(delRequest)
	if err != nil && !strings.Contains(err.Error(), "No redirects") {
		log.Error("Error in connecting to Bosh", err)
		return nil, errors.New(sberrors.ErrBoshConnect)
	}

	taskId, err := getTaskId(resp)
	if err != nil {
		return nil, err
	}

	log.Infof("Successfully initiated delete deployment:%s. Task Id is: %s", deploymentName, taskId)
	return c.GetTask(taskId)
}

func (c *boshHttpClient) GetTask(taskId string) (*schema.Task, error) {
	log.Debug("In GetTask")
	url := fmt.Sprintf("%s%s%s", c.boshDetails.BoshDirectorUrl, "/tasks/", taskId)
	resp, err := c.httpClient.Get(url)
	if err != nil && !strings.Contains(err.Error(), "No redirects") {
		log.Error("Error in connecting to Bosh", err)
		return nil, err
	}
	log.Debug("Received response from Bosh")
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("Not found")
	}
	if resp.StatusCode != http.StatusOK {
		log.Errorf("Non OK status code from BOSH: %d", resp.StatusCode)
		return nil, errors.New(fmt.Sprintf("Non OK status code from BOSH: %d", resp.StatusCode))
	}

	task := schema.Task{}
	err = json.NewDecoder(resp.Body).Decode(&task)
	if err != nil {
		log.Error("Error in decoding response from Bosh", err)
		return nil, err
	}

	return &task, nil
}

func getTaskId(resp *http.Response) (string, error) {
	taskUrl := resp.Header.Get("Location")
	if taskUrl == "" {
		log.Error("Invalid response from Bosh")
		return "", errors.New(sberrors.ErrBoshInvalidResponse)
	}

	split := strings.Split(taskUrl, "/")
	taskId := split[len(split)-1]
	if taskId == "" {
		log.Error("Invalid response from Bosh")
		return "", errors.New(sberrors.ErrBoshInvalidResponse)
	}

	return taskId, nil
}
