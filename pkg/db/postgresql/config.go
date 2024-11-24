package postgresql

import (
	"time"
)

type Config struct {
	Host           string
	Port           uint16
	Database       string
	User           string
	Password       string
	ConnectTimeout time.Duration
}

const defaultConnectTimeout = 5 * time.Second

func (c Config) ToConnectionString() string {
	return generateConnectionString(c)
}

func NewConfigForLocalhost(database string, user string, password string) Config {
	return Config{
		Host:           "localhost",
		Port:           5432,
		Database:       database,
		User:           user,
		Password:       password,
		ConnectTimeout: defaultConnectTimeout,
	}
}

func NewConfigForDockerContainer(database string, user string, password string) Config {
	return Config{
		// https://stackoverflow.com/questions/68173651/connecting-to-a-localhost-postgres-database-from-within-a-docker-container
		Host:           "172.17.0.1",
		Port:           5432,
		Database:       database,
		User:           user,
		Password:       password,
		ConnectTimeout: defaultConnectTimeout,
	}
}
