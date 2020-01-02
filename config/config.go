package config

import (
	"io/ioutil"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// fusion fields
type fusionStruct struct {
	APIKey  string        `yaml:"api_key"`
	URL     string        `yaml:"api_url"`
	Timeout time.Duration `yaml:"timeout"`
}

// server fields
type serverStruct struct {
	Address string `yaml:"listen"`
}

// kafka fields
type kafkaStruct struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}

// Config structure for server
type Config struct {
	Fusion fusionStruct `yaml:"fusion"`
	Server serverStruct `yaml:"server"`
	Kafka  kafkaStruct  `yaml:"kafka"`
}

// ParseYamlFile the config file
func ParseYamlFile(filename string, c *Config) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, c)
}
