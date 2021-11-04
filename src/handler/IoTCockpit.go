package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"log"

	"github.com/tidwall/gjson"
	helper "github.com/uneidel/sapleonardo/helper"
)

//https://c6d4df73-58c3-414f-b854-b4237a4bfc9a.eu10.cp.iot.sap/c6d4df73-58c3-414f-b854-b4237a4bfc9a/iot/core/api/v1/doc/swagger#
//https://developers.sap.com/tutorials/iot-cf-create-device-model-api.html
type IoTCockpit struct {
	httpclient         http.Client
	jwtToken           string
	Username           string
	Password           string
	TenantId           string
	IoTServiceCFAPIURL string
	Debug              bool
}

func (c *IoTCockpit) GetAllGateways() error {
	url := fmt.Sprintf("%s/iot/core/api/v1/tenant/%s/gateways", c.IoTServiceCFAPIURL, c.TenantId)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("cannot create new request: %s", err)
	}
	resp, err := c.ExecuteRequest(request)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read body: %s", err)
	}
	if c.Debug {
		c.prettyPrint(body)
	}
	return nil

}

func (c *IoTCockpit) GetAllSensors() error {
	url := fmt.Sprintf("%s/iot/core/api/v1/tenant/%s/sensors", c.IoTServiceCFAPIURL, c.TenantId)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("cannot create new request: %s", err)
	}
	resp, err := c.ExecuteRequest(request)
	if err != nil {
		return fmt.Errorf("cannot execute request: %s", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read body: %s", err)
	}
	c.prettyPrint(body)
	return nil
}
func (c *IoTCockpit) DeleteSensor(sensorId string) error {
	log.Printf("deleting sensor: %s", sensorId)
	if sensorId == "" {
		return fmt.Errorf("sensor ID is empty")
	}
	url := fmt.Sprintf("%s/iot/core/api/v1/tenant/%s/sensors/%s", c.IoTServiceCFAPIURL, c.TenantId, sensorId)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("cannot create new request: %s", err)
	}
	resp, err := c.ExecuteRequest(request)
	if err != nil {
		return fmt.Errorf("cannot execute request: %s", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("cannot delete logger")
	}
	return nil
}

func (c *IoTCockpit) GetSingleSensor(sensorId string) error {
	url := fmt.Sprintf("%s/iot/core/api/v1/tenant/%s/sensors/%s", c.IoTServiceCFAPIURL, c.TenantId, sensorId)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("cannot create new request: %s", err)
	}
	_, err = c.ExecuteRequest(request) //TODO: ExecuteRequest checks for status code 200. nothing to do here.
	if err != nil {
		return fmt.Errorf("cannot execute request: %s", err)
	}

	return nil
}

func (c *IoTCockpit) CreateNewSensor(ipv6 string, deviceId string, sensortypeId string) (string, error) {
	s := IoTCockpitNewSensor{}
	s.AlternateID = fmt.Sprintf("Sensor_%s", ipv6)
	s.DeviceID = deviceId
	s.Name = fmt.Sprintf("SENSOR_%s", ipv6)
	s.SensorTypeID = sensortypeId
	payload, err := json.Marshal(s)
	if err != nil {
		return "", fmt.Errorf("cannot marshal payload: %s", err)
	}
	url := fmt.Sprintf("%s/iot/core/api/v1/tenant/%s/sensors", c.IoTServiceCFAPIURL, c.TenantId)
	requestbody := bytes.NewBuffer(payload)
	request, err := http.NewRequest(http.MethodPost, url, requestbody)
	if err != nil {
		return "", fmt.Errorf("cannot create new request: %s", err)
	}
	resp, err := c.ExecuteRequest(request)
	if err != nil {
		return "", fmt.Errorf("cannot execute request: %s", err)
	}
	if resp.StatusCode == 409 {
		return "", fmt.Errorf("sensor already exists")
	}

	if resp.StatusCode == 200 {
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("cannot read response body: %v", err)
		}

		res := &IoTCockpitNewSensor{}
		err = json.Unmarshal(b, res)
		if err != nil {
			return "", fmt.Errorf("cannot umarshal response body: %v", err)
		}
		return res.ID, nil
	}

	return "", errors.New("unknown error")
}
func (c *IoTCockpit) GetSensorIdbyAlternateId(alternateId string) (string, error) {
	filter := url.QueryEscape(fmt.Sprintf("alternateId eq '%s'", alternateId))
	qurl := fmt.Sprintf("%s/iot/core/api/v1/tenant/%s/sensors?filter=%s", c.IoTServiceCFAPIURL, c.TenantId, filter)
	request, err := http.NewRequest(http.MethodGet, qurl, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create new request: %s", err)
	}
	resp, err := c.ExecuteRequest(request)
	if err != nil {
		return "", fmt.Errorf("cannot execute request: %s", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("sap call not successfull")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read body: %s", err)
	}
	fmt.Println(string(body))
	var sensor []map[string]interface{}
	err = json.Unmarshal(body, &sensor)
	if err != nil {
		return "", fmt.Errorf("cannot unmarshal body: %s", err)
	}

	if len(sensor) == 0 {
		return "", fmt.Errorf("response did not contain id")
	}

	id, ok := sensor[0]["id"]
	if !ok {
		return "", fmt.Errorf("response does not contain id")
	}
	id2, ok := id.(string)
	if !ok {
		return "", fmt.Errorf("cannot assert id to string")
	}

	return id2, nil
}
func (c *IoTCockpit) GetAllDevices() error {
	url := fmt.Sprintf("%s/iot/core/api/v1/tenant/%s/devices", c.IoTServiceCFAPIURL, c.TenantId)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("cannot create new request: %s", err)
	}
	resp, err := c.ExecuteRequest(request)
	if err != nil {
		return fmt.Errorf("cannot execute request: %s", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read body: %s", err)
	}
	c.prettyPrint(body)
	return nil
}
func (c *IoTCockpit) GetDeviceIdbyName(devicename string) (string, error) {
	filter := url.QueryEscape(fmt.Sprintf("name eq '%s'", devicename))
	url := fmt.Sprintf("%s/iot/core/api/v1/tenant/%s/devices?filter=%s", c.IoTServiceCFAPIURL, c.TenantId, filter)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create new request: %s", err)
	}
	resp, err := c.ExecuteRequest(request)
	if err != nil {
		return "", fmt.Errorf("cannot execute request: %s", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("sap call not successfull")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read body: %s", err)
	}
	var sensor []map[string]interface{}
	err = json.Unmarshal(body, &sensor)
	if err != nil {
		return "", fmt.Errorf("cannot unmarshal body: %s", err)
	}

	if len(sensor) == 0 {
		return "", fmt.Errorf("response did not contain id")
	}

	id, ok := sensor[0]["id"]
	if !ok {
		return "", fmt.Errorf("response does not contain id")
	}
	id2, ok := id.(string)
	if !ok {
		return "", fmt.Errorf("cannot assert id to string")
	}

	return id2, nil

}
func (c *IoTCockpit) DeleteDevice(deviceId string) error {
	log.Printf("deleting device: %s", deviceId)
	if deviceId == "" {
		return fmt.Errorf("deviceID is empty")
	}

	url := fmt.Sprintf("%s/iot/core/api/v1/tenant/%s/devices/%s", c.IoTServiceCFAPIURL, c.TenantId, deviceId)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("cannot create new request: %s", err)
	}
	resp, err := c.ExecuteRequest(request)
	if err != nil {
		return fmt.Errorf("cannot execute request: %s", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("cannot delete device")
	}

	return nil
}
func (c *IoTCockpit) CreateNewDevice(ipv6 string, gatewayId string) (string, error) {
	i := IoTCockpitNewDevice{}
	i.AlternateID = helper.RemoveColon(ipv6)
	i.GatewayID = gatewayId
	i.Name = fmt.Sprintf("DEVICE_%s", ipv6)
	payload, err := json.Marshal(i)
	if err != nil {
		return "", fmt.Errorf("cannot marshal payload: %s", err)
	}
	url := fmt.Sprintf("%s/iot/core/api/v1/tenant/%s/devices", c.IoTServiceCFAPIURL, c.TenantId)
	requestbody := bytes.NewBuffer(payload)
	request, err := http.NewRequest(http.MethodPost, url, requestbody)
	if err != nil {
		return "", fmt.Errorf("cannot create new request: %s", err)
	}
	resp, err := c.ExecuteRequest(request)
	if err != nil {
		return "", fmt.Errorf("cannot execute request: %s", err)
	}
	if resp.StatusCode == 409 {
		return "", fmt.Errorf("device already exists")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read body: %s", err)

	}
	//c.prettyPrint(body)
	var res map[string]interface{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return "", fmt.Errorf("cannot unmarshal body: %s", err)
	}
	deviceId, err := helper.GetStringFromMap(res, "id")
	if err != nil {
		return "", fmt.Errorf("cannot get string id from map: %s", err)
	}
	return deviceId, nil
}

func (c *IoTCockpit) GetAllSensorTypes() error {
	url := fmt.Sprintf("%s/iot/core/api/v1/tenant/%s/sensorTypes", c.IoTServiceCFAPIURL, c.TenantId)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("cannot create new request: %s", err)
	}
	resp, err := c.ExecuteRequest(request)
	if err != nil {
		return fmt.Errorf("cannot execute request: %s", err)
	}
	log.Printf("ResponseCode: %v", resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read body: %s", err)

	}
	c.prettyPrint(body)
	return nil
}
func (c *IoTCockpit) CreateSensorTypebyFile(filename string) (string, error) {
	file, err := c.readJson(filename)
	if err != nil {
		return "", err
	}
	return c.CreateSensorType(file)

}
func (c *IoTCockpit) CreateSensorType(payload []byte) (string, error) {
	url := fmt.Sprintf("%s/iot/core/api/v1/tenant/%s/sensorTypes", c.IoTServiceCFAPIURL, c.TenantId)

	requestbody := bytes.NewBuffer(payload)
	request, err := http.NewRequest(http.MethodPost, url, requestbody)
	if err != nil {
		return "", fmt.Errorf("cannot create new request: %s", err)
	}
	resp, err := c.ExecuteRequest(request)
	if err != nil {
		return "", fmt.Errorf("cannot execute request: %s", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read body: %s", err)
	}
	value := gjson.Get(string(body), "id")

	return value.String(), nil
}
func (c *IoTCockpit) prettyPrint(jsonstr []byte) {
	buf := new(bytes.Buffer)
	json.Indent(buf, []byte(jsonstr), "", "  ")
	log.Printf("Body: %v", buf)
}
func (c *IoTCockpit) GetAllCapabilities() ([]Capability, []byte) {
	capabilities := []Capability{}

	//("/iot/core/api/v1/tenant/766292537/capabilities"
	url := fmt.Sprintf("%s/iot/core/api/v1/tenant/%s/capabilities", c.IoTServiceCFAPIURL, c.TenantId)
	log.Printf("Requesting url: %s", url)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("cannot create request: %s", err)
		return nil, nil
	}
	resp, err := c.ExecuteRequest(request)
	if err != nil {
		log.Printf("cannot execute request: %s", err)
		return nil, nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("cannot read body: %s", err)
		return nil, nil
	}

	err = json.Unmarshal(body, &capabilities)
	if err != nil {
		log.Printf("cannot unmarshal body: %s", err)
		return nil, nil
	}

	return capabilities, body
}

func (c *IoTCockpit) GetCapabilityByName(name string) (*Capability, error) {
	Capabilities, _ := c.GetAllCapabilities()
	if Capabilities == nil {
		return nil, errors.New("cannot get capabilities")
	}

	for _, a := range Capabilities {
		if a.Name == name {
			return &a, nil
		}
	}
	return nil, errors.New("not found")
}

func (c *IoTCockpit) GetCapabilityById(id string) (*Capability, error) {
	Capabilities, _ := c.GetAllCapabilities()
	if Capabilities == nil {
		return nil, errors.New("cannot get capabilities")
	}

	for _, a := range Capabilities {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, errors.New("not found")
}

func (c *IoTCockpit) CreateCapabilitybyFile(filename string, debug bool) (string, error) {

	file, err := c.readJson(filename)
	if err != nil {
		return "", err
	}
	return c.CreateCapability(file)

}
func (c *IoTCockpit) CreateCapability(payload []byte) (string, error) {
	url := fmt.Sprintf("%s/iot/core/api/v1/tenant/%s/capabilities", c.IoTServiceCFAPIURL, c.TenantId)

	requestbody := bytes.NewBuffer(payload)
	request, err := http.NewRequest(http.MethodPost, url, requestbody)
	if err != nil {
		return "", fmt.Errorf("cannot create new request: %s", err)
	}
	resp, err := c.ExecuteRequest(request)
	if err != nil {
		return "", fmt.Errorf("cannot execute request: %s", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read body: %s", err)
	}
	value := gjson.Get(string(body), "id")
	//c.logger.Infof("CapabilityId: %s", value)

	return value.String(), nil
}
func Checkerr(err error) {
	if err != nil {
		panic(err)
	}
}
func (c *IoTCockpit) readJson(path string) ([]byte, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read json file: %s", err)
	}
	return byteValue, nil
}
func (c *IoTCockpit) Init() {
	c.initHttpClient()
}

func (c *IoTCockpit) initHttpClient() {

	c.httpclient = http.Client{}
	log.Println("HttpClient initialized.")
}

//DEPRECATED DO NOT USE!
func (c *IoTCockpit) executeRequest(r *http.Request) (*http.Response, error) {
	c.prepareRequest(r)
	resp, err := c.httpclient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("cannot send request: %s", err)
	}

	if resp.StatusCode != 200 {
		log.Printf("%s resulted in %v", r.URL, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("cannot read body: %s", err)
		} else {
			log.Printf("Err Body: %v", string(body))
		}
		err = errors.New("Request resulted in " + string(resp.StatusCode))
		return resp, err
	} else {
		return resp, nil
	}
}
func (c *IoTCockpit) prepareRequest(http *http.Request) {
	//http.Header.Add("Authorization", "Bearer "+c.jwtToken)
	http.Header.Add("Authorization", "Basic "+basicAuth(c.Username, c.Password))
	http.Header.Add("Content-Type", "application/json")
	//c.logger.Info("Headers added.")
}

func (c *IoTCockpit) ExecuteRequest(r *http.Request) (*http.Response, error) {

	c.prepareRequest(r)
	resp, err := c.httpclient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("cannot execute request: %s", err)
	}

	if resp.StatusCode != 200 {
		log.Printf("%s resulted in %v", r.URL, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("cannot read body: %s", err)
		} else {
			log.Printf("err Body: %v", string(body))
		}
		return resp, fmt.Errorf("request resulted in %d", resp.StatusCode)
	}
	return resp, nil

}
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

type Capability struct {
	ID          string `json:"id"`
	AlternateID string `json:"alternateId"`
	Name        string `json:"name"`
	Properties  []struct {
		Name     string `json:"name"`
		DataType string `json:"dataType"`
	} `json:"properties"`
}

type SensorType struct {
	Capabilities []struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"capabilities"`
	Name string `json:"name"`
}
type IoTCockpitCapability struct {
	Name        string                 `json:"name"`
	AlternateID string                 `json:"alternateId"`
	Properties  []IoTCockpitProperties `json:"properties"`
}
type IoTCockpitProperties struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
}
type IoTCockpitSensorType struct {
	Capabilities []IoTCockpitSensorTypeCapabilities `json:"capabilities"`
	Name         string                             `json:"name"`
}
type IoTCockpitSensorTypeCapabilities struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

/*  IOT Cockpit New Device */
type IoTCockpitNewDevice struct {
	AlternateID      string             `json:"alternateId"`
	CustomProperties []CustomProperties `json:"customProperties"`
	GatewayID        string             `json:"gatewayId"`
	Name             string             `json:"name"`
	Sensors          []Sensors          `json:"sensors"`
}
type CustomProperties struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type Sensors struct {
	AlternateID      string             `json:"alternateId"`
	CustomProperties []CustomProperties `json:"customProperties"`
	Name             string             `json:"name"`
	SensorTypeID     string             `json:"sensorTypeId"`
}

/* End IoT Cockpit New Device */
/* Begin IoT Cocktpit New Sensor */
type IoTCockpitNewSensor struct {
	ID               string             `json:"id,omitempty"`
	AlternateID      string             `json:"alternateId"`
	CustomProperties []CustomProperties `json:"customProperties"`
	DeviceID         string             `json:"deviceId"`
	Name             string             `json:"name"`
	SensorTypeID     string             `json:"sensorTypeId"`
}

/* End IoT Cockpit New Sensor */

/* Helper */
