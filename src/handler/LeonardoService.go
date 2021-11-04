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
	"strconv"
	"strings"
	"time"

	"log"

	"github.com/tidwall/gjson"
)

type ThingStruct struct {
	ExternalID  string `json:"_externalId"`
	AlternateID string `json:"_alternateId"`
	Name        string `json:"_name"`
	Description struct {
		En string `json:"en"`
	} `json:"_description"`
	Thingtemplate []string `json:"thingTemplate"`
	ThingType     []string `json:"_thingType"`
	ObjectGroup   string   `json:"_objectGroup"`
}
type ThingCreateResult struct {
	ID          string `json:"_id"`
	ExternalID  string `json:"_externalId"`
	AlternateID string `json:"_alternateId"`
	Name        string `json:"_name"`
	Description struct {
		En string `json:"en"`
	} `json:"_description"`
	ThingType         []string  `json:"_thingType"`
	CreatedByUser     string    `json:"CreatedByUser"`
	CreatedTime       time.Time `json:"CreatedTime"`
	LastChangedByUser string    `json:"LastChangedByUser"`
	LastChangedTime   time.Time `json:"LastChangedTime"`
	ObjectGroup       string    `json:"_objectGroup"`
}

type Leonardo struct {
	authTokenHost  string
	ClientId       string
	ClientSecret   string
	jwtToken       string
	URL_APPIOT_MDS string
	PackageUrl     string
	ThingUrl       string
	MappingUrl     string
	AuthUrl        string
	TenantId       string
	Debug          bool
	httpclient     http.Client
}

func (l *Leonardo) Init() {

	l.initHttpClient()
	l.PackageUrl = "https://config-package-sap.cfapps.eu10.hana.ondemand.com/Package/v1"
	l.ThingUrl = "https://config-thing-sap.cfapps.eu10.hana.ondemand.com:443/ThingConfiguration/v1"
	l.MappingUrl = "https://tm-data-mapping.cfapps.eu10.hana.ondemand.com/v1"
	l.URL_APPIOT_MDS = "https://appiot-mds.cfapps.eu10.hana.ondemand.com"
}

func (l *Leonardo) initHttpClient() {

	l.httpclient = http.Client{}
	//l.log.Printf("HttpClient initialized.")
}

func (l *Leonardo) GetAllMappingsIds() error {

	url := fmt.Sprintf("%s/mappings/mappingIds", l.MappingUrl)
	r, _ := http.NewRequest("GET", url, nil)
	resp, _ := l.ExecuteRequest(r)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("\n\n", string(body))
	return nil
}

//https://help.sap.com/viewer/080fabc6cae6423fb45fca7752adb61e/1907b/en-US/63f25e40276c4a05bbbc0a69ba5f862e.html
func (l *Leonardo) GetMapping(id string) error {

	url := fmt.Sprintf("%s/Mappings/%s", l.MappingUrl, id)
	r, _ := http.NewRequest("GET", url, nil)
	resp, _ := l.ExecuteRequest(r)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("\n\n", string(body))
	return nil
}

func (l *Leonardo) CreateMapping(jsonstr []byte) error {
	url := fmt.Sprintf("%s/Mappings", l.MappingUrl)
	payload := bytes.NewBuffer(jsonstr)

	r, _ := http.NewRequest("POST", url, payload)

	resp, err := l.ExecuteRequest(r)
	if err != nil {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Body: %v", body)
		log.Printf("Cannot execute Request. %v", err)
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("Body: %v", string(body))
	return nil
}

func (l *Leonardo) CreateThingType(packageName string, thingTypeName string, propertysets []string) error {
	url := fmt.Sprintf("%s/Packages('%s')/ThingTypes", l.ThingUrl, packageName)
	tt := ThingType{}
	tt.Name = fmt.Sprintf("%s:%s", packageName, thingTypeName)
	d := Descriptions{}
	d.Description = "Created by service"
	d.LanguageCode = "en"
	tt.Descriptions = append(tt.Descriptions, d)
	p := PropertySets{}
	for _, a := range propertysets {
		p.Name = a
		p.PropertySetType = fmt.Sprintf("%s:%s", packageName, a)
		tt.PropertySets = append(tt.PropertySets, p)

	}
	jsonstr, _ := json.Marshal(tt)
	payload := bytes.NewBuffer(jsonstr)
	r, _ := http.NewRequest("POST", url, payload)

	resp, err := l.ExecuteRequest(r)
	Checkerr(err)
	if resp.StatusCode != 201 {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Create ThingType Body: %v", string(body))
		return errors.New(fmt.Sprintf("StatusCode: %v returned", resp.StatusCode))
	}

	return nil
}

func (l *Leonardo) GetAllThingTypes() error {
	url := fmt.Sprintf("%s/ThingTypes?$format=json", l.ThingUrl)
	r, _ := http.NewRequest("GET", url, nil)
	resp, _ := l.ExecuteRequest(r)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("\n\n", string(body))
	return nil
}

func (l *Leonardo) DeleteThingType(thingName string) error {
	_, etag, err := l.GetThingType(thingName)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/ThingTypes('%s')", l.ThingUrl, thingName)
	log.Printf("Url: %s", url)
	r, _ := http.NewRequest("DELETE", url, nil)
	r.Header.Add("If-Match", etag)
	resp, err := l.ExecuteRequest(r)

	if err != nil || resp.StatusCode != 204 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Body: %s", string(body))
		return errors.New("Cannot execute Request.")
	}
	defer resp.Body.Close()
	return nil
}

func (l *Leonardo) GetThingType(thingName string) ([]byte, string, error) {
	url := fmt.Sprintf("%s/ThingTypes('%s')?$format=json", l.ThingUrl, thingName)
	r, _ := http.NewRequest("GET", url, nil)
	resp, _ := l.ExecuteRequest(r)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	etag := resp.Header["Etag"]
	if l.Debug {
		log.Printf("Body: %s", string(body))
	}
	return body, etag[0], nil
}

func (l *Leonardo) GetAllThings() {
	url := fmt.Sprintf("%s%s", l.URL_APPIOT_MDS, "/Things")
	r, _ := http.NewRequest("GET", url, nil)
	resp, _ := l.ExecuteRequest(r)

	defer resp.Body.Close()
	fmt.Printf("Headers: %v", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("\n\n", string(body))

}

func (l *Leonardo) CreateThing(externalId string,
	alternateId string,
	name string,
	description string,
	objectGroupId string,
	customer string,
	thingTemplate []string) (*ThingCreateResult, error) {

	data := ThingStruct{ExternalID: externalId, AlternateID: alternateId, Name: name, Thingtemplate: thingTemplate, ObjectGroup: objectGroupId}
	data.Description.En = "v12"
	data.ThingType = thingTemplate
	jsonstr, err := json.Marshal(data)

	if err != nil {
		log.Printf("Cannot convert To Json %v", err)
	}
	//fmt.Printf("\nJson: %s", string(jsonstr))
	payload := bytes.NewBuffer(jsonstr)

	url := fmt.Sprintf("%s%s", l.URL_APPIOT_MDS, "/Things")
	r, _ := http.NewRequest("POST", url, payload)

	resp, err := l.ExecuteRequest(r)
	if err != nil {
		log.Printf("Cannot execute Request. %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	//fmt.Printf("Headers: %v", resp.Header)
	location := resp.Header["Location"]
	if resp.StatusCode != 201 {
		return nil, errors.New("Something went wrong with StatusCode: " + string(resp.StatusCode))
	}
	request, err := http.NewRequest(http.MethodGet, location[0], nil)
	resp, err = l.ExecuteRequest(request)
	body, _ := ioutil.ReadAll(resp.Body)
	if l.Debug {
		log.Printf("Body from Location: %v", string(body))
	}
	thingresult := ThingCreateResult{}
	json.Unmarshal(body, &thingresult)
	return &thingresult, nil
}

func (l Leonardo) GetThing(id string) (ThingCreateResult, error) {

	log.Printf("Get ThingId : %s", id)

	url := fmt.Sprintf("%s('%s')", l.URL_APPIOT_MDS, id)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Error building Request.  %v", err)
		return ThingCreateResult{}, err
	}
	resp, err := l.ExecuteRequest(request)
	body, _ := ioutil.ReadAll(resp.Body)
	thing := ThingCreateResult{}
	json.Unmarshal(body, &thing)
	return thing, nil
}

func (l Leonardo) DeleteThing(ctxID string, id string) error {
	log.Printf("%s: delete ThingId : %s", ctxID, id)
	l.refreshJWTToken()

	url := fmt.Sprintf("%s/Things('%s')", l.URL_APPIOT_MDS, id)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("cannot create new request: %s", err)
	}
	resp, err := l.ExecuteRequest(request)
	if err != nil {
		return fmt.Errorf("cannot execute request: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 204 {
		return nil
	}
	return fmt.Errorf("cannot delete thing")
}

func (l Leonardo) ExecuteRequest(r *http.Request) (*http.Response, error) {
	l.refreshJWTToken()
	l.prepareRequest(r)
	resp, err := l.httpclient.Do(r)
	if err != nil {
		log.Printf("Cannot send Request.%v", err)
	}
	return resp, nil

}

func (l *Leonardo) prepareRequest(http *http.Request) {
	http.Header.Add("Authorization", "Bearer "+l.jwtToken)
	//	l.log.Printf("Token: %v", l.jwtToken)
	http.Header.Add("Accept", "application/json")
	http.Header.Add("Content-Type", "application/json")
	//l.logger.Info("Headers added.")
}
func (l *Leonardo) basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
func (l *Leonardo) refreshJWTToken() {
	apiUrl := l.AuthUrl

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", l.ClientId)
	data.Set("client_secret", l.ClientSecret)
	data.Set("response_type", "token")

	u, err := url.ParseRequestURI(apiUrl)
	if err != nil {
		log.Printf("ParseRequest Uri %v", err)
	}

	u.Path = "/oauth/token"
	urlStr := u.String() // "https://api.com/user/"

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode())) // URL-encoded payload

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	r.Header.Add("Accept-Charset", "utf-8")
	resp, _ := client.Do(r)

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal([]byte(body), &result)
	tokenString := result["access_token"].(string)

	l.jwtToken = tokenString
}
func (l *Leonardo) DeletePackage(packageName string) error {
	//https://config-package-sap.cfapps.eu10.hana.ondemand.com/Package/v1
	_, etag, _ := l.GetPackage(packageName)
	url := fmt.Sprintf("%s%s('%s')", l.PackageUrl, "/Packages", packageName)
	log.Printf("Url: %s", url)
	r, _ := http.NewRequest("DELETE", url, nil)
	r.Header.Add("if-match", etag)
	resp, err := l.ExecuteRequest(r)
	if err != nil || resp.StatusCode != 204 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Body: %s", string(body))
		log.Printf("Cannot execute Request.%v", err)
		return err
	}
	return nil
}

func (l *Leonardo) CreatePackage(payload []byte) error {

	p := bytes.NewBuffer(payload)

	url := fmt.Sprintf("%s%s", l.PackageUrl, "/Packages")
	r, _ := http.NewRequest("POST", url, p)

	resp, err := l.ExecuteRequest(r)
	if err != nil {
		log.Printf("Cannot execute Request. %v", err)
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 201 {
		log.Printf("StatusCode: %v -  Body: %s", resp.StatusCode, string(body))
		return errors.New(fmt.Sprintf("Call failed with StatusCode: %d", resp.StatusCode))
	}
	return nil
}
func (l *Leonardo) GetAllPackages() error {

	url := "https://config-package-sap.cfapps.eu10.hana.ondemand.com/Package/v1/Packages"
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Error building Request.%v", err)
		return err
	}
	resp, err := l.ExecuteRequest(request)
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	log.Printf("Package:%s", string(body))
	return nil
}
func (l Leonardo) GetPackage(packageName string) ([]byte, string, error) {

	url := fmt.Sprintf("%s/Package('%s')?$format=json", l.PackageUrl, packageName)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Error building Request.%v", err)
		return nil, "", err
	}
	resp, err := l.ExecuteRequest(request)
	body, _ := ioutil.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		log.Printf("Body:%s", string(body))
		return nil, "", errors.New("Cannot Process Request")
	}

	etag := resp.Header["Etag"]
	return body, etag[0], nil
}

func (l *Leonardo) readJson(path string) ([]byte, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	return byteValue, nil

}
func (l *Leonardo) GetAllPropertySetTypes(packagename string) error {

	url := fmt.Sprintf("%s/Packages('%s')/PropertySetTypes?$format=json", l.ThingUrl, packagename)
	r, _ := http.NewRequest("GET", url, nil)
	resp, _ := l.ExecuteRequest(r)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("\n\n", string(body))
	return nil
}
func (l *Leonardo) DeletePropertySet(packagename string, psname string) error {
	_, etag, err := l.GetPropertySet(packagename, psname)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/PropertySetTypes('%s:%s')?$expand=Description", l.ThingUrl, packagename, psname)
	r, _ := http.NewRequest("DELETE", url, nil)
	r.Header.Add("if-match", etag)
	resp, err := l.ExecuteRequest(r)
	if err != nil || resp.StatusCode != 204 {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("\n\n", string(body))
		return errors.New("cannot execute request")
	}
	return nil
}
func (l *Leonardo) GetPropertySet(packagename string, psname string) ([]byte, string, error) {
	//"https://config-thing-sap.cfapps.eu10.hana.ondemand.com:443/ThingConfiguration/v1"
	url := fmt.Sprintf("%s/PropertySetTypes('%s:%s')", l.ThingUrl, packagename, psname)

	r, _ := http.NewRequest("GET", url, nil)

	resp, err := l.ExecuteRequest(r)
	if err != nil || resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("\n\n", string(body))
		return nil, "", errors.New("cannot execute request")
	}

	etag := resp.Header["Etag"]

	body, _ := ioutil.ReadAll(resp.Body)
	return body, etag[0], nil
}
func (l *Leonardo) CreatePropertySet(packagename string, jsonstr []byte) (string, error) {

	// Do Post
	payload := bytes.NewBuffer(jsonstr)

	url := fmt.Sprintf("%s/Packages('%s')/PropertySetTypes", l.ThingUrl, packagename)
	//l.log.Printf("Url: %s", url)
	request, _ := http.NewRequest("POST", url, payload)
	resp, err := l.ExecuteRequest(request)
	Checkerr(err)
	body, _ := ioutil.ReadAll(resp.Body)
	if l.Debug {
		log.Printf("Body: %s", string(body))
	}
	value := gjson.Get(string(body), "d.Name")
	return value.String(), nil
}

func (l *Leonardo) GetLeonardoType(input interface{}) (string, error) {

	switch input.(type) {
	case nil:
		return "", errors.New("nil")
	case int32:
		return "Numeric", nil
	case bool:
		return "Boolean", nil
	case string:
		str := fmt.Sprintf("%v", input)

		_, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return "String", nil
		}
		return "DateTime", nil
	case float64:
		if (float64(int(input.(float64))) - input.(float64)) == 0 { // Check if possible Integer
			return "Numeric", nil
		}
		return "NumericFlexible", nil
	default:
		return "", errors.New("unknown")
	}
}

type PropertySet struct {
	Name         string         `json:"Name"`
	DataCategory string         `json:"DataCategory"`
	Descriptions []Descriptions `json:"Descriptions"`
	Properties   []Properties   `json:"Properties"`
}
type Descriptions struct {
	LanguageCode string `json:"LanguageCode"`
	Description  string `json:"Description"`
}
type Properties struct {
	Name           string         `json:"Name"`
	Descriptions   []Descriptions `json:"Descriptions"`
	Type           string         `json:"Type"`
	PropertyLength string         `json:"PropertyLength"`
	QualityCode    string         `json:"QualityCode"`
	UnitOfMeasure  string         `json:"UnitOfMeasure"`
}

type ThingType struct {
	Name         string         `json:"Name"`
	Descriptions []Descriptions `json:"Descriptions"`
	PropertySets []PropertySets `json:"PropertySets"`
}

type PropertySets struct {
	Name            string         `json:"Name"`
	PropertySetType string         `json:"PropertySetType"`
	Descriptions    []Descriptions `json:"Descriptions"`
}
type PackageDefinition struct {
	Name         string         `json:"Name"`
	Scope        string         `json:"Scope"`
	Descriptions []Descriptions `json:"Descriptions"`
}

type MappingDefinition struct {
	Name        string             `json:"name"`
	Description MappingDescription `json:"description"`
	ThingTypeID string             `json:"thingTypeId"`
	Mappings    []MappingMappings  `json:"mappings"`
}
type MappingDescription struct {
	En string `json:"en"`
}
type PropertyMappings struct {
	CapabilityPropertyID string `json:"capabilityPropertyId"`
	NpstPropertyID       string `json:"npstPropertyId"`
}
type Measures struct {
	CapabilityID       string             `json:"capabilityId"`
	NamedPropertySetID string             `json:"namedPropertySetId"`
	PropertyMappings   []PropertyMappings `json:"propertyMappings"`
}
type MappingPropertyMappings struct {
	CapabilityPropertyID string `json:"capabilityPropertyId"`
	NpstRefPropertyID    string `json:"npstRefPropertyId"`
}
type TargetValues struct {
	CapabilityID       string             `json:"capabilityId"`
	NamedPropertySetID string             `json:"namedPropertySetId"`
	PropertyMappings   []PropertyMappings `json:"propertyMappings"`
}
type MappingMappings struct {
	SensorTypeID string         `json:"sensorTypeId"`
	Measures     []Measures     `json:"measures"`
	TargetValues []TargetValues `json:"targetValues"`
}
