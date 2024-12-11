package config

import (
	"encoding/json"
	"github.com/ghodss/yaml"
	"log"
	"os"
)

type Metrics struct {
	System  bool `yaml:"system"`
	Sensors bool `yaml:"sensors"`
	Power   bool `yaml:"power"`
	Sel     bool `yaml:"sel"`
	Storage bool `yaml:"storage"`
	Memory  bool `yaml:"memory"`
	Network bool `yaml:"network"`
}

type Basic struct {
	BindIp string `yaml:"bindIp"`
	Port   string `yaml:"port"`
}

type Hosts struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Config struct {
	Hosts   map[string]Hosts `yaml:"hosts"`
	Basic   Basic            `yaml:"basic"`
	Metrics Metrics          `yaml:"metrics"`
}

func Init(path string) Config {
	var config Config

	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
		return config
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error unmarshalling YAML: %v", err)
		return config
	}
	return config
}

type Map struct {
	Ip   string `yaml:"ip"`
	Name string `yaml:"name"`
}

func GetMapOfNameAndIp(path, target string) string {
	var mapOfIpAndName []Map
	content, err := os.ReadFile(path)
	if err != nil {
		log.Println("Error reading file:", err)
		return mapOfIpAndName[0].Name
	}
	err = json.Unmarshal(content, &mapOfIpAndName)
	if err != nil {
		log.Println("Error unmarshalling YAML:", err)
		return mapOfIpAndName[0].Name
	}
	for _, item := range mapOfIpAndName {
		if item.Ip == target {
			return item.Name
		}
	}
	return mapOfIpAndName[0].Name
}
