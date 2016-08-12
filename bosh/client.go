package bosh

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	sberrors "github.com/atulkc/fabric-service-broker/errors"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("bosh")

type Client interface {
	CreateDeployment(manifest Manifest) (*Task, error)
	DeleteDeployment(deploymentName string) (*Task, error)
	GetTask(taskId string) (*Task, error)
	GetVmIps(deploymentName string) (map[string][]string, error)
}

type boshHttpClient struct {
	httpClient  *http.Client
	boshDetails *Details
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

func NewBoshHttpClient(boshDetails *Details) Client {
	return &boshHttpClient{
		boshDetails: boshDetails,
		httpClient:  newHttpClient(true),
	}
}

func (c *boshHttpClient) CreateDeployment(manifest Manifest) (*Task, error) {
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

func (c *boshHttpClient) DeleteDeployment(deploymentName string) (*Task, error) {
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

func (c *boshHttpClient) GetTask(taskId string) (*Task, error) {
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

	task := Task{}
	err = json.NewDecoder(resp.Body).Decode(&task)
	if err != nil {
		log.Error("Error in decoding response from Bosh", err)
		return nil, err
	}

	return &task, nil
}

func parseVMIpsFromResponse(response *http.Response) (map[string][]string, error) {
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error("Error reading from response", err)
		return nil, err
	}

	vmDetails := strings.Split(string(data), "\n")
	vmIpsMap := make(map[string][]string)
	for _, vmDetail := range vmDetails {
		if strings.Trim(vmDetail, " ") == "" {
			log.Debug("Skipping empty line")
			continue
		}
		vmDetailMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(vmDetail), &vmDetailMap)
		if err != nil {
			log.Error("Error in unmarshaling vm details", err)
			continue
		}
		jobName := vmDetailMap["job_name"].(string)
		log.Debugf("Parsing job %s", jobName)
		vmIps := vmDetailMap["ips"].([]interface{})
		if len(vmIps) < 1 {
			log.Errorf("No IPs for job %s", jobName)
			continue
		}
		vmIp := vmIps[0].(string)
		log.Debugf("IP for job:%s is %s", jobName, vmIp)

		jobIps, exists := vmIpsMap[jobName]
		if !exists {
			log.Debugf("No entry for %s in map...making one", jobName)
			jobIps = make([]string, 0)
		}
		jobIps = append(jobIps, vmIp)
		log.Debugf("IPs for job:%s is %s", jobName, jobIps)
		vmIpsMap[jobName] = jobIps
	}

	return vmIpsMap, nil
}

func (c *boshHttpClient) GetVmIps(deploymentName string) (map[string][]string, error) {
	log.Debug("In GetVMIps")

	url := fmt.Sprintf("%s/deployments/%s/vms?format=full", c.boshDetails.BoshDirectorUrl, deploymentName)
	resp, err := c.httpClient.Get(url)
	if err != nil && !strings.Contains(err.Error(), "No redirects") {
		log.Error("Error in connecting to Bosh", err)
		return nil, err
	}
	log.Debug("Received response from Bosh")

	taskId, err := getTaskId(resp)
	if err != nil {
		return nil, err
	}
	log.Debugf("Initiated get vm details operation. Task ID:%s", taskId)

	taskOutputUrl := fmt.Sprintf("%s/tasks/%s/output?type=result", c.boshDetails.BoshDirectorUrl, taskId)
	attempts := 1
	log.Debug("Wait before querying the task output")
	time.Sleep(500 * time.Millisecond)
	for attempts <= 3 {
		log.Debugf("Attempt number: %d to get VM details", attempts)
		taskOutputResponse, err := c.httpClient.Get(taskOutputUrl)
		if err != nil {
			log.Error("Error in connecting to Bosh", err)
			return nil, err
		}
		if taskOutputResponse.StatusCode == http.StatusOK {
			return parseVMIpsFromResponse(taskOutputResponse)
		}
		if taskOutputResponse.StatusCode != http.StatusNoContent {
			log.Errorf("Error in getting deployment details. Status code: %d.", taskOutputResponse.StatusCode)
			return nil, errors.New("Error in getting deployment details")
		}

		time.Sleep(time.Duration(attempts) * time.Second)
		attempts++
	}
	log.Errorf("Could not get deployment details after %d attemts", attempts)
	return nil, errors.New("Max retries exceeded in getting deployment details from Bosh")
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
