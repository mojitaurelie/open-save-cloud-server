package config

import (
	"flag"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Configuration struct {
	Database DatabaseConfiguration `yaml:"database"`
	Features FeaturesConfiguration `yaml:"features"`
	Cache    string                `yaml:"cache"`
}

type DatabaseConfiguration struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type FeaturesConfiguration struct {
	AllowRegister bool `yaml:"allow_register"`
}

var currentConfig *Configuration

func init() {
	path := flag.String("config", "./config.yml", "Set the configuration file path")
	flag.Parse()
	configYamlContent, err := os.ReadFile(*path)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(configYamlContent, &currentConfig)
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

func Database() *DatabaseConfiguration {
	return &currentConfig.Database
}

func Features() *FeaturesConfiguration {
	return &currentConfig.Features
}

func Cache() string {
	return currentConfig.Cache
}
