package config

import (
	"go.uber.org/zap"
	yaml "gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Mode        string           `yaml:"mode"`
	ServiceName string           `yaml:"serviceName"`
	Secret      string           `yaml:"secret"`
	Server      ServerConfig     `yaml:"server"`
	DB          DBConfig         `yaml:"db"`
	Prometheus  PrometheusConfig `yaml:"prometheus"`
}

type ServerConfig struct {
	Port     int    `yaml:"port"`
	GRPCPort int    `yaml:"grpc_port"`
	Scheme   string `yaml:"scheme"`
	Domain   string `yaml:"domain"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type PrometheusConfig struct {
	Port int `yaml:"port"`
}

func MustLoad(configPath string) Config {
	conf := Config{}

	data, err := os.ReadFile(configPath)
	if err != nil {
		panic("failed to read config: " + err.Error())
	}

	if err = yaml.Unmarshal(data, &conf); err != nil {
		panic("failed to unmarshal config: " + err.Error())
	}

	zap.L().Info(
		"Load configuration from yaml",
		zap.String("path", configPath),
	)
	return conf
}
