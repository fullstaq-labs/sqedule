package cli

import (
	"errors"

	"github.com/spf13/viper"
)

type Config struct {
	ServerBaseURL     string
	BasicAuthUser     string
	BasicAuthPassword string
	Debug             bool
}

func LoadConfigFromViper(viper *viper.Viper) Config {
	return Config{
		ServerBaseURL:     viper.GetString("server-base-url"),
		BasicAuthUser:     viper.GetString("basic-auth-user"),
		BasicAuthPassword: viper.GetString("basic-auth-password"),
		Debug:             viper.GetBool("debug"),
	}
}

func (config Config) RequireServerConfig() error {
	if len(config.ServerBaseURL) == 0 {
		return errors.New("Configuration required: server-base-url")
	}

	return nil
}
