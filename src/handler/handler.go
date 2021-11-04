package handler

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

var (
	iotcockpit IoTCockpit
	leonardo   Leonardo
)

func LoadConfig() Config {
	filename := "config.yml"
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	config := Config{}
	err = yaml.Unmarshal(source, &config)
	return config
}

func GetConfig() (IoTCockpit, Leonardo) {
	log.Printf("Loading Config")
	c := LoadConfig()
	iotcockpit = IoTCockpit{}
	iotcockpit.IoTServiceCFAPIURL = c.IotCockpit.IoTServiceCFAPIURL
	iotcockpit.Username = c.IotCockpit.Username
	iotcockpit.Password = c.IotCockpit.Password
	iotcockpit.TenantId = c.IotCockpit.TenantId
	iotcockpit.Init()
	leonardo = Leonardo{}
	leonardo.AuthUrl = c.Leonardo.AuthUrl
	leonardo.ClientId = c.Leonardo.ClientId
	leonardo.ClientSecret = c.Leonardo.ClientSecret
	leonardo.TenantId = c.Leonardo.TenantId
	leonardo.Init()
	log.Printf("%v", c)

	return iotcockpit, leonardo
}

type Config struct {
	IotCockpit struct {
		IoTServiceCFAPIURL string `yaml:"IoTServiceCFAPIURL"`
		Username           string `yaml:"Username"`
		Password           string `yaml:"Password"`
		TenantId           string `yaml:"TenantId"`
	} `yaml:"Iotcockpit"`
	Leonardo struct {
		AuthUrl      string `yaml:"AuthUrl"`
		ClientId     string `yaml:"ClientId"`
		ClientSecret string `yaml:"ClientSecret"`
		TenantId     string `yaml:"TenantId"`
	} `yaml:"Leonardo"`
}
