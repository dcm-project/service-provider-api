package config

import (
	"github.com/kelseyhightower/envconfig"
)

var singleConfig *Config = nil

type Config struct {
	Database *dbConfig
	Service  *svcConfig
}

type dbConfig struct {
	Type     string `envconfig:"DB_TYPE" default:"pgsql"`
	Hostname string `envconfig:"DB_HOST" default:"localhost"`
	Port     string `envconfig:"DB_PORT" default:"5433"`
	Name     string `envconfig:"DB_NAME" default:"service-provider"`
	User     string `envconfig:"DB_USER" default:"admin"`
	Password string `envconfig:"DB_PASS" default:"adminpass"`
}

type svcConfig struct {
	Address  string `envconfig:"DCM_ADDRESS" default:":8081"`
	BaseUrl  string `envconfig:"DCM_BASE_URL" default:"https://localhost:8081"`
	LogLevel string `envconfig:"DCM_LOG_LEVEL" default:"info"`
}

func New() (*Config, error) {
	if singleConfig == nil {
		singleConfig = new(Config)
		if err := envconfig.Process("", singleConfig); err != nil {
			return nil, err
		}
	}
	return singleConfig, nil
}
