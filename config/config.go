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
	ListenAddr string        `yaml:"listen_addr"`
	AdminAddr  string        `yaml:"admin_addr"`
	IDService  string        `yaml:"id_service"`
	Timeout    time.Duration `yaml:"timeout"`
	Cache      string        `yaml:"cache"`
	Switch     bool          `yaml:"tls_switch"`
	Perm       string        `yaml:"tls_perm"`
	Key        string        `yaml:"tls_key"`
	Ca         string        `yaml:"tls_ca"`
}

// redis fields
type redisStruct struct {
	Passwd string `yaml:"passwd"`
	Addr   string `yaml:"addr"`
	DB     int    `yaml:"db"`
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
	Switch  bool     `yaml:"tls_switch"`
	Perm    string   `yaml:"tls_perm"`
	Key     string   `yaml:"tls_key"`
	Ca      string   `yaml:"tls_ca"`
}

// log fields
type logStruct struct {
	OutputLevel        string `yaml:"output_level"`
	OutputPath         string `yaml:"output_path"`
	RotationPath       string `yaml:"rotation_path"`
	RotationMaxSize    int    `yaml:"rotation_max_size"`
	RotationMaxAge     int    `yaml:"rotation_max_age"`
	RotationMaxBackups int    `yaml:"rotation_max_backups"`
	JSONEncoding       bool   `yaml:"json_encoding"`
}

// Config structure for server
type Config struct {
	Fusion fusionStruct `yaml:"fusion"`
	Server serverStruct `yaml:"server"`
	Cache  cacheStruct  `yaml:"cache"`
	Redis  redisStruct  `yaml:"redis"`
	Kafka  kafkaStruct  `yaml:"kafka"`
	Log    logStruct    `yaml:"log"`
}

// ParseYamlFile the config file
func ParseYamlFile(filename string, c *Config) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, c)
}
