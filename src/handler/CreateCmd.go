package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	helper "github.com/uneidel/sapleonardo/helper"
)

const (
	SEPARATOR = "_"
)

type Createcmd struct {
	Version  string `required`
	Filename string `required`
}

func (cmd *Createcmd) Run() error {
	GetConfig()
	log.Printf("Opening %s", cmd.Filename)
	file, err := helper.ReadJson(cmd.Filename)
	if err != nil {
		panic(err)
	}
	jo := map[string]interface{}{}
	json.Unmarshal(file, &jo)
	x := helper.FlattenAndOmit(jo, true)

	outfile, _ := json.Marshal(x)
	ioutil.WriteFile("message_omitted.json", outfile, 0644)
	sensId, capId, _ := CreateIoTCockpitStuff(iotcockpit, cmd.Version, x)
	log.Printf("SensorID: %s -  CapId: %s", sensId, capId)
	log.Printf("Sleeping for a while.")

	CreateLeonardoStuff(leonardo, cmd.Version, outfile, sensId, capId)

	log.Printf("Done.")
	return nil

}
func CreateLeonardoStuff(leonardo Leonardo, Version string, file []byte, sensId string, capId string) {
	log.Printf("Creating Leonardo Portion")
	log.Printf("Creating Package Definition")
	packagedef, err := CreateLeonardoPackageDefinition(leonardo.TenantId, Version)
	Checkerr(err)
	log.Printf("Creating Package")
	err = leonardo.CreatePackage(packagedef)
	Checkerr(err)
	packagename := fmt.Sprintf("%s.%s", "sw.dev", Version)
	propertyNames := []string{}
	ps, _ := CreatePropertySets(leonardo.TenantId, Version, file)
	leonardo.Debug = true
	for _, a := range ps {
		jsonstr, _ := json.Marshal(a)
		psName, err := leonardo.CreatePropertySet(packagename, jsonstr)
		if len(psName) == 0 {
			continue
		}
		s := strings.Split(psName, ":")
		if len(s) != 2 {
			log.Printf("Error splitting : %v", psName)
		}
		propertyNames = append(propertyNames, s[1])
		log.Printf("PropertysetName :%v", psName)
		Checkerr(err)

	}
	leonardo.Debug = false
	log.Printf("Create ThingType")
	err = leonardo.CreateThingType(packagename, fmt.Sprintf("ThingType_%s", Version), propertyNames)
	Checkerr(err)
	thingtypeId := fmt.Sprintf("%s:ThingType_%s", packagename, Version)
	log.Printf("THingTypeId: %s", thingtypeId)
	//leonardo.GetAllMappingsIds()
	//leonardo.GetMapping("ad2c3b48-1eeb-43bd-8d14-9918a173976c")
	log.Printf("Create Mapping")
	jsonstr, _ := CreateMapping(Version, thingtypeId, sensId, capId, file)
	//log.Printf("%v", string(jsonstr))
	log.Printf("Apply Mapping")
	leonardo.CreateMapping(jsonstr)
}
func CreateIoTCockpitStuff(iotcockpit IoTCockpit, Version string, x map[string]interface{}) (string, string, error) {
	log.Printf("Creating IoTCockpit Stuff")
	payload, err := CreateIoTCockpitCapability(Version, x)
	if err != nil {

		panic(err)
	}
	err = ioutil.WriteFile("./tmp/cap.json", payload, 0644)
	if err != nil {
		log.Printf("Cannot write to tmp")

		panic(err)
	}

	capId, err := iotcockpit.CreateCapability(payload)
	if err != nil {
		panic(err)
	}
	//log.Infof("Created CapId: %v", capId)

	log.Printf("Create SensorType")
	capIds := []string{capId}
	sensorpayload, _ := CreateIoTCockpitSensortype(Version, capIds)
	sensId, _ := iotcockpit.CreateSensorType(sensorpayload)
	//log.Infof("SensorId: %s", sensId)

	return sensId, capId, nil

}
func CreateMapping(version string, thingtypeId string, sensortypeId string, capId string, rawjson []byte) ([]byte, error) {

	p := MappingDefinition{}
	p.Name = fmt.Sprintf("Mapping_%s", version)
	p.Description = MappingDescription{}
	p.Description.En = fmt.Sprintf("Mapping for Thingtype_%s", version)
	p.ThingTypeID = thingtypeId
	name := "main"
	m := MappingMappings{}
	m.SensorTypeID = sensortypeId
	measures := Measures{}
	measures.CapabilityID = capId
	measures.NamedPropertySetID = fmt.Sprintf("%s", name)
	// Creating Main

	var j map[string]interface{}
	err := json.Unmarshal(rawjson, &j)
	Checkerr(err)
	for k, v := range j {
		switch v.(type) {
		case map[string]interface{}:
			sub := v.(map[string]interface{})
			x, _ := CreateSubMeasures(capId, sensortypeId, k, sub)
			m.Measures = append(m.Measures, x)

		default:
			pm := PropertyMappings{}
			pm.CapabilityPropertyID = fmt.Sprintf("%s", k)
			pm.NpstPropertyID = k
			measures.PropertyMappings = append(measures.PropertyMappings, pm)

		}
	}
	m.Measures = append(m.Measures, measures)
	p.Mappings = append(p.Mappings, m)
	jsonstr, err := json.Marshal(p)
	return jsonstr, err
}
func CreateSubMeasures(capId string, sensortypeId string, key string, kv map[string]interface{}) (Measures, error) {

	measures := Measures{}
	measures.NamedPropertySetID = key
	measures.CapabilityID = capId
	for k, _ := range kv {
		y := PropertyMappings{}

		y.CapabilityPropertyID = fmt.Sprintf("%s%s%s", key, SEPARATOR, k)
		y.NpstPropertyID = k
		measures.PropertyMappings = append(measures.PropertyMappings, y)
	}

	return measures, nil
}

// Current only one nested level supported
func CreatePropertySets(tenantid string, version string, rawjson []byte) ([]PropertySet, error) {

	psets := []PropertySet{}
	main := PropertySet{}
	main.Name = fmt.Sprintf("%s.%s:%s", tenantid, version, "main")
	main.DataCategory = "TimeSeriesData"
	d := Descriptions{}
	d.LanguageCode = "en"
	d.Description = fmt.Sprintf("%s Sensor Parameter", "main")
	main.Descriptions = append(main.Descriptions, d)
	var p map[string]interface{}
	err := json.Unmarshal(rawjson, &p)
	Checkerr(err)
	for k, v := range p {
		switch v.(type) {
		case map[string]interface{}:
			//log.Printf("Key %s is map %v", k, v)
			sub := v.(map[string]interface{})
			ps, err := CreateSubPropertySet(tenantid, version, k, sub)
			Checkerr(err)
			psets = append(psets, ps)

		default:
			//log.Printf("Key: %s with Value %v No Map added directly", k, v)
			p := Properties{}
			p.Name = k
			ld := Descriptions{}
			ld.Description = fmt.Sprintf("%s Sensor Values", k)
			ld.LanguageCode = "en"
			p.Descriptions = append(p.Descriptions, ld)
			p.Type, _ = leonardo.GetLeonardoType(v)
			p.QualityCode = "0"
			p.PropertyLength = ""
			p.UnitOfMeasure = ""
			main.Properties = append(main.Properties, p)
		}

	}

	psets = append(psets, main)
	return psets, nil
}
func CreateSubPropertySet(tenantid string, version string, nodename string, kv map[string]interface{}) (PropertySet, error) {
	sub := PropertySet{}
	sub.Name = fmt.Sprintf("%s.%s:%s", tenantid, version, nodename)
	sub.DataCategory = "TimeSeriesData"
	d := Descriptions{}
	d.LanguageCode = "en"
	d.Description = fmt.Sprintf("%s Sensor Parameter", nodename)
	sub.Descriptions = append(sub.Descriptions, d)
	for k, v := range kv {
		p := Properties{}
		p.Name = k
		ld := Descriptions{}
		ld.Description = fmt.Sprintf("%s Sensor Values", k)
		ld.LanguageCode = "en"
		p.Descriptions = append(p.Descriptions, ld)
		p.Type, _ = leonardo.GetLeonardoType(v)
		p.QualityCode = "0"
		p.PropertyLength = ""
		p.UnitOfMeasure = ""
		sub.Properties = append(sub.Properties, p)
	}
	return sub, nil
}

func CreateLeonardoPackageDefinition(tenantid string, version string) ([]byte, error) {
	p := PackageDefinition{}
	p.Name = fmt.Sprintf("%s.%s", tenantid, version)
	p.Scope = "tenant"

	d := Descriptions{}
	d.LanguageCode = "en"
	d.Description = fmt.Sprintf("Package Definition for %s", p.Name)

	p.Descriptions = append(p.Descriptions, d)

	jsonstr, err := json.Marshal(p)
	return jsonstr, err
}

func CreateIoTCockpitSensortype(version string, capIds []string) ([]byte, error) {
	i := IoTCockpitSensorType{}
	i.Name = fmt.Sprintf("SensorType_%s", version)
	for _, a := range capIds {
		x := IoTCockpitSensorTypeCapabilities{}
		x.ID = a
		x.Type = "measure"
		i.Capabilities = append(i.Capabilities, x)
	}
	jsonstr, err := json.Marshal(i)
	return jsonstr, err
}

func CreateIoTCockpitCapability(version string, x map[string]interface{}) ([]byte, error) {
	i := IoTCockpitCapability{}
	i.Name = fmt.Sprintf("%s_%s", "DataCapability", version)
	i.AlternateID = i.Name
	for k, v := range x {
		valtype, _ := GetIoTCockpitType(v)
		fmt.Printf("%v %v : %v with Type %s \n", "", k, v, valtype)
		if len(valtype) > 0 {
			p := IoTCockpitProperties{}
			p.Name = k
			p.DataType = valtype
			i.Properties = append(i.Properties, p)
		}
	}

	jsonstr, _ := json.Marshal(i)

	return jsonstr, nil
}

func GetIoTCockpitType(input interface{}) (string, error) {

	switch input.(type) {
	case nil:
		return "", errors.New("nil")
	case int32:
		return "integer", nil
	case bool:
		return "boolean", nil
	case string:
		str := fmt.Sprintf("%v", input)
		//log.Infof("-->str: %s", str)
		_, err := time.Parse(time.RFC3339, str)
		if err != nil {
			//log.Infof("-----> error %v", err)
			//TODO Check for Date
			return "string", nil
		}
		//log.Infof("-----> DATE parsed.")
		return "date", nil
	case float64:
		if (float64(int(input.(float64))) - input.(float64)) == 0 { // Check if possible Integer
			return "integer", nil
		}
		return "double", nil
	default:
		return "", errors.New("unknown")
	}

}
