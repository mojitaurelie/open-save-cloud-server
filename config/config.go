package config

import (
	"flag"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Configuration struct {
	Server   ServerConfiguration   `yaml:"server"`
	Database DatabaseConfiguration `yaml:"database"`
	Features FeaturesConfiguration `yaml:"features"`
	Path     PathConfiguration     `yaml:"path"`
}

type PathConfiguration struct {
	Cache   string `yaml:"cache"`
	Storage string `yaml:"storage"`
}

type ServerConfiguration struct {
	Port int `yaml:"port"`
}

type DatabaseConfiguration struct {
	Host     string  `yaml:"host"`
	Port     int     `yaml:"port"`
	Username string  `yaml:"username"`
	Password *string `yaml:"password"`
}

type FeaturesConfiguration struct {
	AllowRegister    bool `yaml:"allow_register"`
	PasswordHashCost *int `yaml:"password_hash_cost"`
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
	checkConfig()
}

func checkConfig() {
	if currentConfig.Features.PasswordHashCost == nil {
		currentConfig.Features.PasswordHashCost = new(int)
		*currentConfig.Features.PasswordHashCost = bcrypt.DefaultCost
	} else if *currentConfig.Features.PasswordHashCost < bcrypt.MinCost && *currentConfig.Features.PasswordHashCost > bcrypt.MaxCost {
		log.Fatalf("password_hash_cost is not on the supported range (%d < x < %d)", bcrypt.MinCost, bcrypt.MaxCost)
	}
	if _, err := os.Stat(currentConfig.Path.Storage); err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(currentConfig.Path.Cache); err != nil {
		log.Fatal(err)
	}
}

func Database() *DatabaseConfiguration {
	return &currentConfig.Database
}

func Features() *FeaturesConfiguration {
	return &currentConfig.Features
}

func Path() *PathConfiguration {
	return &currentConfig.Path
}

func Server() *ServerConfiguration {
	return &currentConfig.Server
}
