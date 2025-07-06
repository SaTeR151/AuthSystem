package config

import (
	"os"

	"github.com/sirupsen/logrus"
)

type ServerConfig struct {
	Port string
}

type PostgresqlConfig struct {
	User    string
	Pass    string
	Dbname  string
	Sslmode string
	Port    string
	Host    string
}

func GetServerConfig() ServerConfig {
	var serverConfig ServerConfig
	var ok bool
	serverConfig.Port, ok = os.LookupEnv("SERVER_PORT")
	if !ok {
		logrus.Warn("server port is empty")
	}
	return serverConfig
}

func GetPostresqlConfig() PostgresqlConfig {
	var psqlConfig PostgresqlConfig
	var ok bool
	psqlConfig.User, ok = os.LookupEnv("POSTGRES_USER")
	if !ok {
		logrus.Warn("postgres user is empty")
	}
	psqlConfig.Pass, ok = os.LookupEnv("POSTGRES_PASSWORD")
	if !ok {
		logrus.Warn("postgres password if empty")
	}
	psqlConfig.Dbname, ok = os.LookupEnv("POSTGRES_DB")
	if !ok {
		logrus.Warn("postgres database name if empty")
	}
	psqlConfig.Sslmode, ok = os.LookupEnv("SSLMODE")
	if !ok {
		logrus.Warn("postgres sslmode is empty")
	}
	psqlConfig.Port, ok = os.LookupEnv("POSTGRES_PORT")
	if !ok {
		logrus.Warn("postgres port is empty")
	}
	psqlConfig.Host, ok = os.LookupEnv("POSTGRES_HOST")
	if !ok {
		logrus.Warn("postgres host is empty")
	}
	return psqlConfig
}

func InitLoggerConfig() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	lvl, ok := os.LookupEnv("LOG_LEVEL")

	if !ok {
		lvl = "debug"
	}

	ll, err := logrus.ParseLevel(lvl)
	if err != nil {
		ll = logrus.DebugLevel
	}

	logrus.SetLevel(ll)
}
