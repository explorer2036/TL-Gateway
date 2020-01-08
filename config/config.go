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
	ListenAddr string `yaml:"listen_addr"`
	AdminAddr  string `yaml:"admin_addr"`
}

// cache fields
type cacheStruct struct {
	Routines  int           `yaml:"routines"`
	Buffer    int           `yaml:"buffer"`
	DelayTime time.Duration `yaml:"delay_time"`
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
	Cache  cacheStruct  `yaml:"cache"`
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
